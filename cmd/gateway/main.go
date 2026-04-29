package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EkantBajaj/aegis-pay/internal/idempotency"
	"github.com/EkantBajaj/aegis-pay/internal/routing"
	"github.com/EkantBajaj/aegis-pay/internal/transport"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// 0. Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	// 1. Configuration
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	stripeURL := os.Getenv("STRIPE_URL")
	if stripeURL == "" {
		stripeURL = "http://localhost:8081/stripe/v1/charges"
	}

	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "localhost:19092"
	}

	// 2. Initialize Dependencies
	idemClient := idempotency.NewClient(redisAddr)
	idemMiddleware := idempotency.NewMiddleware(idemClient)

	// Initialize the Routing Layer (Stripe Client with Circuit Breaker)
	stripeClient := routing.NewProviderClient("Stripe-Main", stripeURL)

	// Initialize the Kafka Producer for the Recovery "Slow Path"
	kafkaProducer := transport.NewKafkaProducer([]string{kafkaBrokers}, "failed-transactions")
	defer kafkaProducer.Close()

	// 3. Initialize Fiber App
	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	app.Use(logger.New())

	// 4. Routes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	app.Post("/charge", idemMiddleware, func(c *fiber.Ctx) error {
		var req routing.ChargeRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}

		// Generate a unique ID for this transaction lifecycle
		txID := "tx_" + time.Now().Format("20060102150405")
		
		log.Printf("--- Gateway: Routing charge for user %s to Stripe ---", req.UserID)
		
		// Attempt charge through the "Fast Path" (Stripe)
		resp, err := stripeClient.Charge(c.Context(), req)
		
		if err != nil {
			log.Printf("--- FAST PATH FAILED: %v. Triggering Slow Path (AI Recovery) ---", err)
			
			// 1. Publish to Kafka for the AI Agent to handle
			failureEvent := map[string]interface{}{
				"transaction_id": txID,
				"request":        req,
				"error":          err.Error(),
				"failed_at":      time.Now().Format(time.RFC3339),
			}
			
			if kErr := kafkaProducer.PublishFailure(c.Context(), failureEvent); kErr != nil {
				log.Printf("--- CRITICAL: Failed to publish to Kafka: %v ---", kErr)
				return c.Status(500).JSON(fiber.Map{"error": "internal_system_error"})
			}

			// 2. Return 202 Accepted to the user
			return c.Status(202).JSON(fiber.Map{
				"status": "pending_recovery",
				"message": "Payment is taking longer than usual. We are working on it.",
				"transaction_id": txID,
			})
		}

		// Success Path
		return c.Status(200).JSON(resp)
	})

	// 5. Server Lifecycle & Graceful Shutdown
	go func() {
		log.Printf("Aegis-Pay Gateway starting on :8080")
		if err := app.Listen(":8080"); err != nil {
			log.Panicf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down Aegis-Pay Gateway...")

	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Fatalf("Forceful shutdown: %v", err)
	}

	log.Println("Gateway stopped gracefully.")
}

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EkantBajaj/aegis-pay/internal/idempotency"
	"github.com/EkantBajaj/aegis-pay/internal/routing"
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

	// 2. Initialize Dependencies
	idemClient := idempotency.NewClient(redisAddr)
	idemMiddleware := idempotency.NewMiddleware(idemClient)

	// Initialize the Routing Layer (Stripe Client with Circuit Breaker)
	stripeClient := routing.NewProviderClient("Stripe-Main", stripeURL)

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

		log.Printf("--- Gateway: Routing charge for user %s to Stripe ---", req.UserID)
		
		// Attempt charge through the client (protected by circuit breaker)
		resp, err := stripeClient.Charge(c.Context(), req)
		if err != nil {
			log.Printf("--- FAILURE: Stripe call failed: %v ---", err)
			return c.Status(502).JSON(fiber.Map{
				"error": "payment_failed",
				"detail": err.Error(),
			})
		}

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

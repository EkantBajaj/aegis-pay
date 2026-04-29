package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EkantBajaj/aegis-pay/internal/idempotency"
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
		redisAddr = "localhost:6379" // Default for local dev
	}

	// 2. Initialize Dependencies (Dependency Injection)
	// We inject the Redis client into the Idempotency Middleware
	idemClient := idempotency.NewClient(redisAddr)
	idemMiddleware := idempotency.NewMiddleware(idemClient)

	// 3. Initialize Fiber App
	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	})

	// Add standard logger to see requests in terminal
	app.Use(logger.New())

	// 4. Routes
	
	// Health Check (Public, No Idempotency)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Charge Endpoint (Protected by Idempotency Middleware)
	// Every request here MUST have an 'Idempotency-Key' header
	app.Post("/charge", idemMiddleware, func(c *fiber.Ctx) error {
		log.Println("--- Logic: Processing payment in Handler ---")
		
		// Simulate a provider call (Stripe/Adyen) taking 2 seconds
		time.Sleep(2 * time.Second)

		return c.Status(200).JSON(fiber.Map{
			"status": "success",
			"message": "Payment processed successfully",
			"transaction_id": "tx_" + time.Now().Format("20060102150405"),
		})
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

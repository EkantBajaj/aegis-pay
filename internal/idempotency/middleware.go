package idempotency

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// NewMiddleware creates a Fiber handler that enforces idempotency
func NewMiddleware(client *Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Get the key from the header
		key := c.Get("Idempotency-Key")
		if key == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Idempotency-Key header is required for this operation",
			})
		}

		// 2. Try to acquire the lock
		// TTL of 24 hours is standard for payment idempotency
		acquired, err := client.CheckAndLock(c.Context(), key, 24*time.Hour)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "idempotency check failed"})
		}

		if !acquired {
			// 3. Lock failed. Either it's in progress or we have a result.
			val, err := client.GetResult(c.Context(), key)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to retrieve idempotency state"})
			}

			if val == "IN_PROGRESS" {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{
					"error": "An operation with this key is already in progress",
				})
			}

			// 4. Replay the cached response
			// In a real app, we'd store the status code too. For now, assume 200.
			return c.Status(fiber.StatusOK).SendString(val)
		}

		// 5. Success! Proceed to the actual handler
		err = c.Next()

		// 6. After the handler finishes, save the result if successful
		if err == nil {
			// Capture the response body to cache it
			respBody := c.Response().Body()
			_ = client.SetResult(c.Context(), key, string(respBody), 24*time.Hour)
		}

		return err
	}
}

package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// HealthCheck handles health checks for liveness and readiness
func HealthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "auth-service",
	})
}

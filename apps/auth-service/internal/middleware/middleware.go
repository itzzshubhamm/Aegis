package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const CorrelationIDHeader = "x-correlation-id"

// CorrelationID is a middleware that handles x-correlation-id header
func CorrelationID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get(CorrelationIDHeader)
		if id == "" {
			id = uuid.New().String()
		}
		
		c.Set(CorrelationIDHeader, id)
		c.Locals("correlationId", id)
		
		err := c.Next()
		
		c.Response().Header.Set(CorrelationIDHeader, id)
		return err
	}
}

// Logger is a structured logging middleware using slog
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		startTime := time.Now()
		
		err := c.Next()
		
		duration := time.Since(startTime)
		correlationID, _ := c.Locals("correlationId").(string)
		
		slog.Info("Request completed",
			"correlationId", correlationID,
			"method", c.Method(),
			"url", c.OriginalURL(),
			"statusCode", c.Response().StatusCode(),
			"duration", duration.String(),
		)
		
		return err
	}
}

// Recovery handles panics and returns structured JSON responses
func Recovery() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				correlationID, _ := c.Locals("correlationId").(string)
				slog.Error("Recovery caught panic",
					"correlationId", correlationID,
					"path", c.OriginalURL(),
					"error", r,
				)
				
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"statusCode":    fiber.StatusInternalServerError,
					"timestamp":     time.Now().Format(time.RFC3339),
					"path":          c.OriginalURL(),
					"correlationId": correlationID,
					"message":       "Internal server error",
				})
			}
		}()
		return c.Next()
	}
}

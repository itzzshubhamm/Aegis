package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CorrelationIDHeader = "X-Correlation-ID"

// CorrelationID injects or preserves X-Correlation-ID header in gin.Context
func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(CorrelationIDHeader)
		if id == "" {
			id = uuid.New().String()
		}

		c.Set("correlationId", id)
		c.Header(CorrelationIDHeader, id)

		c.Next()
	}
}

// Logger provides slog structured request logging
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		duration := time.Since(startTime)
		correlationID, _ := c.Get("correlationId")

		if raw != "" {
			path = path + "?" + raw
		}

		slog.Info("Request completed",
			"correlationId", correlationID,
			"method", c.Request.Method,
			"path", path,
			"statusCode", c.Writer.Status(),
			"duration", duration.String(),
			"clientIp", c.ClientIP(),
		)
	}
}

// CORS provides standard cross-origin resource sharing middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Correlation-ID, X-Tenant-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

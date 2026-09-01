package handlers

import (
	"net/http"
	"time"

	"auth-service/internal/db"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	store *db.Store
}

func NewHealthHandler(store *db.Store) *HealthHandler {
	return &HealthHandler{store: store}
}

// HealthCheck handles GET /health endpoint
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	dbStatus := "connected"
	statusCode := http.StatusOK

	if h.store != nil {
		if err := h.store.Ping(); err != nil {
			dbStatus = "disconnected: " + err.Error()
			statusCode = http.StatusServiceUnavailable
		}
	} else {
		dbStatus = "not_configured"
	}

	c.JSON(statusCode, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "aegis-backend",
		"database":  dbStatus,
	})
}

package handlers

import (
	"log/slog"
	"net/http"

	"auth-service/internal/db"
	"auth-service/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TenantHandler struct {
	store *db.Store
}

func NewTenantHandler(store *db.Store) *TenantHandler {
	return &TenantHandler{
		store: store,
	}
}

type CreateTenantRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateTenant creates a new tenant organization
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant payload: " + err.Error()})
		return
	}

	tenant, err := h.store.CreateTenant(c.Request.Context(), req.Name)
	if err != nil {
		slog.Error("Failed to create tenant", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		return
	}

	c.JSON(http.StatusCreated, tenant)
}

// GetTenantByID fetches tenant info by ID, enforcing tenant isolation
func (h *TenantHandler) GetTenantByID(c *gin.Context) {
	tenantIDStr := c.Param("id")
	targetTenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID format"})
		return
	}

	// Resolve tenant strictly from authenticated identity
	authTenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant context not found"})
		return
	}

	// Cross-tenant protection check
	if authTenantID != targetTenantID {
		slog.Warn("Cross-tenant access attempt blocked", "authenticated_tenant", authTenantID, "target_tenant", targetTenantID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Cross-tenant access forbidden"})
		return
	}

	tenant, err := h.store.GetTenantByID(c.Request.Context(), targetTenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

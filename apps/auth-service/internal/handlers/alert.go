package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"auth-service/internal/db"
	"auth-service/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AlertHandler struct {
	store *db.Store
}

func NewAlertHandler(store *db.Store) *AlertHandler {
	return &AlertHandler{store: store}
}

type UpdateAlertStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *AlertHandler) ListAlerts(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant identity not resolved"})
		return
	}

	if h == nil || h.store == nil {
		c.JSON(http.StatusOK, []db.Alert{})
		return
	}

	severity := c.Query("severity")
	status := c.Query("status")
	detectionType := c.Query("detection_type")

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	offsetStr := c.DefaultQuery("offset", "0")
	pageStr := c.Query("page")
	offset := 0
	if pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			offset = (page - 1) * limit
		}
	} else if off, err := strconv.Atoi(offsetStr); err == nil && off >= 0 {
		offset = off
	}

	var severityNull sql.NullString
	if severity != "" {
		severityNull = sql.NullString{String: severity, Valid: true}
	}

	var statusNull sql.NullString
	if status != "" {
		statusNull = sql.NullString{String: status, Valid: true}
	}

	var detectionTypeNull sql.NullString
	if detectionType != "" {
		detectionTypeNull = sql.NullString{String: detectionType, Valid: true}
	}

	alerts, err := h.store.ListAlertsFiltered(c.Request.Context(), db.ListAlertsFilteredParams{
		TenantID:      tenantID,
		Severity:      severityNull,
		Status:        statusNull,
		DetectionType: detectionTypeNull,
		Limit:         int32(limit),
		Offset:        int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list alerts: " + err.Error()})
		return
	}

	if alerts == nil {
		alerts = []db.Alert{}
	}

	c.JSON(http.StatusOK, alerts)
}

func (h *AlertHandler) GetAlertByID(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant identity not resolved"})
		return
	}

	if h == nil || h.store == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store unavailable"})
		return
	}

	idStr := c.Param("id")
	alertID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID format"})
		return
	}

	alert, err := h.store.GetAlertByID(c.Request.Context(), db.GetAlertByIDParams{
		ID:       alertID,
		TenantID: tenantID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, alert)
}

func (h *AlertHandler) UpdateAlertStatus(c *gin.Context) {
	tenantID, ok := middleware.GetTenantID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant identity not resolved"})
		return
	}

	if h == nil || h.store == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store unavailable"})
		return
	}

	idStr := c.Param("id")
	alertID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID format"})
		return
	}

	var req UpdateAlertStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Validate status
	validStatuses := map[string]bool{"OPEN": true, "ACKNOWLEDGED": true, "RESOLVED": true, "CLOSED": true, "INVESTIGATING": true}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
		return
	}

	updatedAlert, err := h.store.UpdateAlertStatus(c.Request.Context(), db.UpdateAlertStatusParams{
		ID:       alertID,
		Status:   req.Status,
		TenantID: tenantID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, updatedAlert)
}

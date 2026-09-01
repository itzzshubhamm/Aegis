package middleware

import (
	"net/http"
	"strings"

	"auth-service/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	ContextUserIDKey    = "userId"
	ContextTenantIDKey  = "tenantId"
	ContextUserEmailKey = "userEmail"
	ContextUserRoleKey  = "userRole"
	ContextClaimsKey    = "claims"
)

// AuthMiddleware validates Authorization Bearer token and injects authenticated user identity & tenant into gin.Context
func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := auth.ValidateToken(parts[1], secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if claims.TokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type, access token required"})
			c.Abort()
			return
		}

		c.Set(ContextClaimsKey, claims)
		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextTenantIDKey, claims.TenantID)
		c.Set(ContextUserEmailKey, claims.Email)
		c.Set(ContextUserRoleKey, claims.Role)

		c.Next()
	}
}

// GetTenantID extracts tenant_id from authenticated request context
func GetTenantID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextTenantIDKey)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// GetUserID extracts user_id from authenticated request context
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextUserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

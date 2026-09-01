package handlers

import (
	"log/slog"
	"net/http"

	"auth-service/internal/auth"
	"auth-service/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	store     *db.Store
	jwtSecret string
}

func NewAuthHandler(store *db.Store, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		store:     store,
		jwtSecret: jwtSecret,
	}
}

type RegisterRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	TenantName string `json:"tenant_name"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register registers a new tenant and first user (tenant admin)
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid registration payload: " + err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Check if user already exists
	existingUser, err := h.store.GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser.ID != uuid.Nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	// Determine tenant name
	tenantName := req.TenantName
	if tenantName == "" {
		tenantName = req.Email + "'s Org"
	}

	// 1. Create Tenant
	tenant, err := h.store.CreateTenant(ctx, tenantName)
	if err != nil {
		slog.Error("Failed to create tenant during registration", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		return
	}

	// 2. Hash Password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		slog.Error("Failed to hash password", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process registration"})
		return
	}

	// 3. Create First User as Tenant Admin/Owner
	user, err := h.store.CreateUser(ctx, db.CreateUserParams{
		TenantID:     tenant.ID,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         "admin",
	})
	if err != nil {
		slog.Error("Failed to create user during registration", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// 4. Generate JWT Pair
	accessToken, refreshToken, err := auth.GenerateTokenPair(user.ID, tenant.ID, user.Email, user.Role, h.jwtSecret)
	if err != nil {
		slog.Error("Failed to generate JWT tokens", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to issue auth tokens"})
		return
	}

	slog.Info("User registered successfully", "user_id", user.ID, "tenant_id", tenant.ID, "email", user.Email)

	c.JSON(http.StatusCreated, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"tenant_id":  user.TenantID,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		},
		"tenant": gin.H{
			"id":   tenant.ID,
			"name": tenant.Name,
		},
		"token":         accessToken,
		"refresh_token": refreshToken,
	})
}

// Login authenticates credentials and returns JWT token pair
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login payload: " + err.Error()})
		return
	}

	ctx := c.Request.Context()

	user, err := h.store.GetUserByEmail(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	accessToken, refreshToken, err := auth.GenerateTokenPair(user.ID, user.TenantID, user.Email, user.Role, h.jwtSecret)
	if err != nil {
		slog.Error("Failed to generate JWT tokens on login", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to issue auth tokens"})
		return
	}

	slog.Info("User logged in successfully", "user_id", user.ID, "tenant_id", user.TenantID)

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"tenant_id":  user.TenantID,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.CreatedAt,
		},
		"token":         accessToken,
		"refresh_token": refreshToken,
	})
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid refresh payload: " + err.Error()})
		return
	}

	claims, err := auth.ValidateToken(req.RefreshToken, h.jwtSecret)
	if err != nil || claims.TokenType != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	ctx := c.Request.Context()
	user, err := h.store.GetUserByID(ctx, claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User no longer exists"})
		return
	}

	accessToken, refreshToken, err := auth.GenerateTokenPair(user.ID, user.TenantID, user.Email, user.Role, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to issue auth tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":         accessToken,
		"refresh_token": refreshToken,
	})
}

// Logout acknowledges user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

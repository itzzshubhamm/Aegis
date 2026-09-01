package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"auth-service/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-jwt-secret-key"

	userID := uuid.New()
	tenantID := uuid.New()
	accessToken, refreshToken, err := auth.GenerateTokenPair(userID, tenantID, "user@tenant.com", "admin", secret)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	router := gin.New()
	router.Use(AuthMiddleware(secret))
	router.GET("/protected", func(c *gin.Context) {
		resUserID, ok1 := GetUserID(c)
		resTenantID, ok2 := GetTenantID(c)

		if !ok1 || resUserID != userID {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID mismatch"})
			return
		}
		if !ok2 || resTenantID != tenantID {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Tenant ID mismatch"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Test 1: Missing Authorization Header
	req1, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for missing auth header, got %d", w1.Code)
	}

	// Test 2: Invalid Header Format
	req2, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req2.Header.Set("Authorization", "InvalidHeaderFormat")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid header format, got %d", w2.Code)
	}

	// Test 3: Valid Access Token
	req3, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req3.Header.Set("Authorization", "Bearer "+accessToken)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusOK {
		t.Errorf("Expected 200 for valid access token, got %d, body: %s", w3.Code, w3.Body.String())
	}

	// Test 4: Refresh Token used on Access-protected Endpoint
	req4, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req4.Header.Set("Authorization", "Bearer "+refreshToken)
	w4 := httptest.NewRecorder()
	router.ServeHTTP(w4, req4)
	if w4.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 when passing refresh token to access route, got %d", w4.Code)
	}
}

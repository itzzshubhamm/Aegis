package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestPasswordHashing(t *testing.T) {
	password := "SecretP@ssword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if !CheckPasswordHash(password, hash) {
		t.Errorf("CheckPasswordHash failed for correct password")
	}

	if CheckPasswordHash("WrongPassword", hash) {
		t.Errorf("CheckPasswordHash succeeded for wrong password")
	}
}

func TestTokenGenerationAndValidation(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	email := "test@example.com"
	role := "admin"
	secret := "test-secret-key-123"

	accessToken, refreshToken, err := GenerateTokenPair(userID, tenantID, email, role, secret)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	// Validate Access Token
	accessClaims, err := ValidateToken(accessToken, secret)
	if err != nil {
		t.Fatalf("ValidateToken access token failed: %v", err)
	}
	if accessClaims.UserID != userID || accessClaims.TenantID != tenantID || accessClaims.Email != email || accessClaims.Role != role {
		t.Errorf("Access token claims mismatch")
	}
	if accessClaims.TokenType != "access" {
		t.Errorf("Expected token type 'access', got %s", accessClaims.TokenType)
	}

	// Validate Refresh Token
	refreshClaims, err := ValidateToken(refreshToken, secret)
	if err != nil {
		t.Fatalf("ValidateToken refresh token failed: %v", err)
	}
	if refreshClaims.TokenType != "refresh" {
		t.Errorf("Expected token type 'refresh', got %s", refreshClaims.TokenType)
	}

	// Test Invalid Secret
	_, err = ValidateToken(accessToken, "wrong-secret")
	if err == nil {
		t.Errorf("Expected error validating token with wrong secret, got nil")
	}
}

func TestExpiredToken(t *testing.T) {
	secret := "test-secret-key-123"
	expiredClaims := Claims{
		UserID:    uuid.New(),
		TenantID:  uuid.New(),
		Email:     "expired@example.com",
		Role:      "analyst",
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	tokenStr, err := tokenObj.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to sign expired token: %v", err)
	}

	_, err = ValidateToken(tokenStr, secret)
	if err == nil {
		t.Errorf("Expected error for expired token, got nil")
	}
}

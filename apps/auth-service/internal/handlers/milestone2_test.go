package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/db"
	"auth-service/internal/kafka"
	"auth-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func setupTestEnvironment(t *testing.T) (*db.Store, *kafka.Producer, *gin.Engine, string) {
	gin.SetMode(gin.TestMode)
	cfg := config.LoadConfig()

	store, err := db.NewStore(cfg.DatabaseURL)
	if err != nil {
		t.Skipf("Skipping DB-dependent integration test: database connection failed: %v", err)
	}

	// Apply migration
	migrationPath := filepath.Join("..", "..", "db", "migrations", "000001_init_schema.up.sql")
	if _, err := os.Stat(migrationPath); err != nil {
		migrationPath = filepath.Join("db", "migrations", "000001_init_schema.up.sql")
	}
	if err := db.InitSchema(store.DB(), migrationPath); err != nil {
		t.Logf("Migration warning: %v", err)
	}

	producer := kafka.NewProducer(cfg.KafkaBrokers, "aegis.events")

	r := gin.New()
	r.Use(middleware.CorrelationID())

	authHandler := NewAuthHandler(store, cfg.JWTSecret)
	tenantHandler := NewTenantHandler(store)
	eventHandler := NewEventHandler(store, producer)

	r.POST("/api/v1/auth/register", authHandler.Register)
	r.POST("/api/v1/auth/login", authHandler.Login)
	r.POST("/api/v1/auth/refresh", authHandler.Refresh)

	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		protected.POST("/auth/logout", authHandler.Logout)
		protected.POST("/tenants", tenantHandler.CreateTenant)
		protected.GET("/tenants/:id", tenantHandler.GetTenantByID)
		protected.POST("/events/ingest", eventHandler.IngestEvent)
	}

	return store, producer, r, cfg.JWTSecret
}

func TestMilestone2Flows(t *testing.T) {
	store, producer, router, _ := setupTestEnvironment(t)
	defer store.Close()
	defer producer.Close()

	uniqueEmail1 := fmt.Sprintf("admin-%d@tenant-a.com", time.Now().UnixNano())
	uniqueEmail2 := fmt.Sprintf("admin-%d@tenant-b.com", time.Now().UnixNano())

	var tokenA, tokenB string
	var tenantIDA, tenantIDB string

	// 1. User Registration - Tenant A
	t.Run("User Registration - Tenant A", func(t *testing.T) {
		payload := map[string]string{
			"email":       uniqueEmail1,
			"password":    "Password123!",
			"tenant_name": "Tenant A Corp",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected 201 Created for registration, got %d. Response: %s", w.Code, w.Body.String())
		}

		var res map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &res)

		tokenA = res["token"].(string)
		tenantInfo := res["tenant"].(map[string]interface{})
		tenantIDA = tenantInfo["id"].(string)

		userInfo := res["user"].(map[string]interface{})
		if userInfo["role"] != "admin" {
			t.Errorf("Expected first registered user to have role 'admin', got %v", userInfo["role"])
		}
	})

	// 2. User Registration - Tenant B
	t.Run("User Registration - Tenant B", func(t *testing.T) {
		payload := map[string]string{
			"email":       uniqueEmail2,
			"password":    "Password123!",
			"tenant_name": "Tenant B Corp",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected 201 Created for registration, got %d. Response: %s", w.Code, w.Body.String())
		}

		var res map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &res)

		tokenB = res["token"].(string)
		tenantInfo := res["tenant"].(map[string]interface{})
		tenantIDB = tenantInfo["id"].(string)
	})

	// 3. Login Verification
	t.Run("User Login - Valid Credentials", func(t *testing.T) {
		payload := map[string]string{
			"email":    uniqueEmail1,
			"password": "Password123!",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK on login, got %d", w.Code)
		}
	})

	// 4. Login Verification - Invalid Credentials
	t.Run("User Login - Invalid Credentials", func(t *testing.T) {
		payload := map[string]string{
			"email":    uniqueEmail1,
			"password": "WrongPassword123",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Expected 401 Unauthorized for wrong password, got %d", w.Code)
		}
	})

	// 5. Tenant Isolation Verification
	t.Run("Tenant Isolation - Tenant A accessing Tenant A info", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/tenants/"+tenantIDA, nil)
		req.Header.Set("Authorization", "Bearer "+tokenA)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 OK when Tenant A accesses Tenant A, got %d", w.Code)
		}
	})

	t.Run("Tenant Isolation - Tenant B accessing Tenant B info", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/tenants/"+tenantIDB, nil)
		req.Header.Set("Authorization", "Bearer "+tokenB)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 OK when Tenant B accesses Tenant B, got %d", w.Code)
		}
	})

	t.Run("Tenant Isolation - Tenant A attempting to access Tenant B info", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/tenants/"+tenantIDB, nil)
		req.Header.Set("Authorization", "Bearer "+tokenA) // Using Tenant A's token

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected 403 Forbidden for cross-tenant access, got %d. Body: %s", w.Code, w.Body.String())
		}
	})

	// 6. Event Ingestion - Valid Payload
	t.Run("Event Ingestion - Valid Payload", func(t *testing.T) {
		payload := map[string]interface{}{
			"event_type": "AUTH_FAILURE",
			"source":     "firewall-edge-01",
			"source_ip":  "192.168.1.100",
			"asset":      "/api/v1/auth/login",
			"severity":   "HIGH",
			"metadata": map[string]interface{}{
				"attempts": 5,
			},
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/events/ingest", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+tokenA)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected 201 Created for event ingestion, got %d. Body: %s", w.Code, w.Body.String())
		}

		var res map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &res)

		if res["tenant_id"] != tenantIDA {
			t.Errorf("Expected event tenant_id to be resolved as Tenant A (%s), got %v", tenantIDA, res["tenant_id"])
		}
	})

	// 7. Event Ingestion - Invalid Payload
	t.Run("Event Ingestion - Invalid Payload", func(t *testing.T) {
		payload := map[string]interface{}{
			// Missing required fields event_type, source, source_ip
			"severity": "LOW",
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/events/ingest", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+tokenA)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request for invalid payload, got %d", w.Code)
		}
	})
}

package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"auth-service/internal/config"
	"auth-service/internal/db"
	"auth-service/internal/handlers"
	"auth-service/internal/kafka"
	"auth-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Gin Engine
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// Middlewares
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.CorrelationID())
	r.Use(middleware.Logger())

	// Initialize DB Store
	var store *db.Store
	if cfg.DatabaseURL != "" {
		var err error
		store, err = db.NewStore(cfg.DatabaseURL)
		if err != nil {
			slog.Error("Failed to initialize database store", "error", err)
		} else {
			defer store.Close()

			// Run auto schema migration if migration script exists
			migrationPath := filepath.Join("db", "migrations", "000001_init_schema.up.sql")
			if _, err := os.Stat(migrationPath); err == nil {
				if err := db.InitSchema(store.DB(), migrationPath); err != nil {
					slog.Warn("Failed to apply DB schema migration", "error", err)
				}
			} else {
				// Try path relative to root directory
				altPath := filepath.Join("apps", "auth-service", "db", "migrations", "000001_init_schema.up.sql")
				if _, err := os.Stat(altPath); err == nil {
					if err := db.InitSchema(store.DB(), altPath); err != nil {
						slog.Warn("Failed to apply DB schema migration from alt path", "error", err)
					}
				}
			}
		}
	}

	// Initialize Kafka Producer
	producer := kafka.NewProducer(cfg.KafkaBrokers, kafka.DefaultTopic)
	defer producer.Close()

	// Initialize Handlers
	healthHandler := handlers.NewHealthHandler(store)
	authHandler := handlers.NewAuthHandler(store, cfg.JWTSecret)
	tenantHandler := handlers.NewTenantHandler(store)
	eventHandler := handlers.NewEventHandler(store, producer)

	// Public Routes
	r.GET("/health", healthHandler.HealthCheck)
	r.GET("/api/v1/health", healthHandler.HealthCheck)

	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
	}

	// Protected Routes (JWT Auth Required)
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		protected.POST("/auth/logout", authHandler.Logout)

		protected.POST("/tenants", tenantHandler.CreateTenant)
		protected.GET("/tenants/:id", tenantHandler.GetTenantByID)

		protected.POST("/events/ingest", eventHandler.IngestEvent)
	}

	// Start Server
	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("Aegis Go Backend Service is starting", "port", cfg.Port, "env", cfg.Env)

	if err := r.Run(addr); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

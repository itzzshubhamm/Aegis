package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"auth-service/internal/config"
	"auth-service/internal/db"
	"auth-service/internal/detection"
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

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

			// Run auto schema migrations
			migrationPath := filepath.Join("db", "migrations")
			if _, err := os.Stat(migrationPath); err == nil {
				if err := db.InitSchema(store.DB(), migrationPath); err != nil {
					slog.Warn("Failed to apply DB schema migration from db/migrations", "error", err)
				}
			} else {
				// Try path relative to workspace root directory
				altPath := filepath.Join("apps", "auth-service", "db", "migrations")
				if _, err := os.Stat(altPath); err == nil {
					if err := db.InitSchema(store.DB(), altPath); err != nil {
						slog.Warn("Failed to apply DB schema migration from alt path", "error", err)
					}
				}
			}
		}
	}

	// Initialize State Tracker (Redis with fallback to InMemory)
	var stateTracker detection.StateTracker
	redisTracker, err := detection.NewRedisStateTracker(cfg.RedisURL)
	if err != nil {
		slog.Warn("Redis connection unavailable, falling back to in-memory state tracker", "error", err)
		stateTracker = detection.NewInMemoryStateTracker()
	} else {
		slog.Info("Successfully connected to Redis for detection state tracking", "url", cfg.RedisURL)
		stateTracker = redisTracker
	}

	// Initialize Detection Engine and Rules
	bruteForceRule := detection.NewBruteForceRule(stateTracker, 3, 5*time.Minute)
	unauthorizedRule := detection.NewUnauthorizedAccessRule()
	honeytokenRule := detection.NewHoneytokenRule()

	engine := detection.NewEngine(bruteForceRule, unauthorizedRule, honeytokenRule)

	// Initialize Kafka Producer
	producer := kafka.NewProducer(cfg.KafkaBrokers, kafka.DefaultTopic)
	defer producer.Close()

	// Initialize Kafka Consumer
	consumer := kafka.NewConsumer(cfg.KafkaBrokers, kafka.DefaultTopic, "aegis-detection-group", engine, store)
	defer consumer.Close()

	// Start background Kafka detection consumer loop
	go consumer.Start(ctx)

	// Initialize Handlers
	healthHandler := handlers.NewHealthHandler(store)
	authHandler := handlers.NewAuthHandler(store, cfg.JWTSecret)
	tenantHandler := handlers.NewTenantHandler(store)
	eventHandler := handlers.NewEventHandler(store, producer)
	alertHandler := handlers.NewAlertHandler(store)
	honeytokenHandler := handlers.NewHoneytokenHandler(store, producer)

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

		// Event APIs
		protected.POST("/events/ingest", eventHandler.IngestEvent)
		protected.GET("/events", eventHandler.ListEvents)
		protected.GET("/events/:id", eventHandler.GetEventByID)

		// Alert APIs
		protected.GET("/alerts", alertHandler.ListAlerts)
		protected.GET("/alerts/:id", alertHandler.GetAlertByID)
		protected.PATCH("/alerts/:id", alertHandler.UpdateAlertStatus)

		// Deception / Honeytoken APIs
		protected.POST("/deception/honeytokens", honeytokenHandler.CreateHoneytoken)
		protected.GET("/deception/honeytokens", honeytokenHandler.ListHoneytokens)
		protected.GET("/deception/honeytokens/:id", honeytokenHandler.GetHoneytokenByID)
		protected.POST("/deception/trigger/:token_id", honeytokenHandler.TriggerHoneytoken)
	}

	// Handle OS signals for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("Aegis Go Backend Service is starting", "port", cfg.Port, "env", cfg.Env)

	go func() {
		if err := r.Run(addr); err != nil {
			slog.Error("Failed to start HTTP server", "error", err)
		}
	}()

	<-quit
	slog.Info("Shutting down Aegis service...")
	cancel()
}

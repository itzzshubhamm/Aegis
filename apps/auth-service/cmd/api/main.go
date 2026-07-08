package main

import (
	"fmt"
	"log/slog"
	"os"

	"auth-service/internal/config"
	"auth-service/internal/handlers"
	"auth-service/internal/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

func main() {
	// Initialize structured JSON logging (like pino)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Load configuration from env variables
	cfg := config.LoadConfig()

	// Initialize Fiber App
	app := fiber.New(fiber.Config{
		AppName: "Aegis Auth Service",
	})

	// Global Security Middlewares
	app.Use(helmet.New())
	app.Use(cors.New())

	// Custom Request Flow Middlewares
	app.Use(middleware.CorrelationID())
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())

	// Health Route
	app.Get("/health", handlers.HealthCheck)

	// Start Server
	addr := fmt.Sprintf(":%d", cfg.Port)
	slog.Info("Auth Service is starting", "port", cfg.Port, "env", cfg.Env)
	
	if err := app.Listen(addr); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

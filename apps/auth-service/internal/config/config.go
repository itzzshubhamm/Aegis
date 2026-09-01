package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         int
	Env          string
	DatabaseURL  string
	RedisURL     string
	KafkaBrokers string
	JWTSecret    string
}

// LoadConfig parses environment variables and returns a Config struct
func LoadConfig() *Config {
	portStr := getEnv("PORT", "4001")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 4001
	}

	return &Config{
		Port:         port,
		Env:          getEnv("NODE_ENV", "development"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://aegis_admin:admin123@localhost:5432/aegis_db?sslmode=disable"),
		RedisURL:     getEnv("REDIS_URL", "localhost:6379"),
		KafkaBrokers: getEnv("KAFKA_BROKERS", "localhost:9092"),
		JWTSecret:    getEnv("JWT_SECRET", "aegis-super-secret-jwt-key-change-in-prod"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

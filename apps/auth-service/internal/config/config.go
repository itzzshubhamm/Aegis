package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        int
	Env         string
	DatabaseURL string
	RedisURL    string
}

// LoadConfig parses environment variables and returns a Config struct
func LoadConfig() *Config {
	portStr := getEnv("PORT", "4001")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 4001
	}

	return &Config{
		Port:        port,
		Env:         getEnv("NODE_ENV", "development"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		RedisURL:    getEnv("REDIS_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	os.Setenv("PORT", "5000")
	defer os.Unsetenv("PORT")

	cfg := LoadConfig()
	if cfg.Port != 5000 {
		t.Errorf("Expected port 5000, got %d", cfg.Port)
	}

	if cfg.DatabaseURL == "" {
		t.Errorf("Expected non-empty default DatabaseURL")
	}
}

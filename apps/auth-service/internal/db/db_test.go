package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPostgresConnectionAndQueries(t *testing.T) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://aegis_admin:admin123@localhost:5432/aegis_db?sslmode=disable"
	}

	store, err := NewStore(dbURL)
	if err != nil {
		t.Fatalf("Failed to open DB connection: %v", err)
	}
	defer store.Close()

	if err := store.Ping(); err != nil {
		t.Skipf("Skipping integration test: PostgreSQL is not reachable: %v", err)
	}

	// Apply migration
	migPath := filepath.Join("..", "..", "db", "migrations", "000001_init_schema.up.sql")
	if err := InitSchema(store.DB(), migPath); err != nil {
		t.Fatalf("Schema migration failed: %v", err)
	}

	// Test sqlc CreateTenant query
	ctx := context.Background()
	tenant, err := store.CreateTenant(ctx, "Test Org Tenant")
	if err != nil {
		t.Fatalf("Failed to create tenant via sqlc query: %v", err)
	}

	if tenant.Name != "Test Org Tenant" {
		t.Errorf("Expected tenant name 'Test Org Tenant', got '%s'", tenant.Name)
	}

	t.Logf("Successfully created tenant with ID: %v", tenant.ID)
}

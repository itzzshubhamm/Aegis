package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
)

// InitSchema executes initial DDL schema migration if tables do not exist
func InitSchema(db *sql.DB, migrationFilePath string) error {
	content, err := os.ReadFile(migrationFilePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", migrationFilePath, err)
	}

	_, err = db.Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration script: %w", err)
	}

	slog.Info("Database schema migration executed successfully")
	return nil
}

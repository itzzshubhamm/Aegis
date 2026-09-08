package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// InitSchema executes DDL schema migrations from a file or directory path
func InitSchema(db *sql.DB, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat migration path %s: %w", path, err)
	}

	var files []string
	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("failed to read migration directory %s: %w", path, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
				files = append(files, filepath.Join(path, entry.Name()))
			}
		}
		sort.Strings(files)
	} else {
		files = append(files, path)
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration script %s: %w", file, err)
		}
		slog.Info("Database migration applied", "file", file)
	}

	slog.Info("All database schema migrations executed successfully")
	return nil
}

package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(dbURL string) (*Store, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		slog.Warn("Database ping failed on startup, will retry during health checks", "error", err)
	} else {
		slog.Info("Successfully connected to PostgreSQL database")
	}

	return &Store{
		Queries: New(db),
		db:      db,
	}, nil
}

func (s *Store) Ping() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	return s.db.Ping()
}

func (s *Store) Close() error {
	if s != nil && s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Store) DB() *sql.DB {
	return s.db
}

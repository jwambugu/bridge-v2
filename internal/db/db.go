package db

import (
	"bridge/internal/config"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

var ErrInvalidDSN = errors.New("db: dsn cannot be empty") // empty dsn

// NewConnection attempt to create a database connection with the provided config.DbURL
func NewConnection() (*sqlx.DB, error) {
	dsn := config.Get[string](config.DbURL, "")
	if dsn == "" {
		return nil, ErrInvalidDSN
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("new db connection: %w", err)
	}
	return db, nil
}

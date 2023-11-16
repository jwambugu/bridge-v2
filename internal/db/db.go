package db

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

var ErrInvalidDSN = errors.New("db: dsn cannot be empty")

type UserTblColumn uint8

const (
	UserUnknown UserTblColumn = iota
	UserEmail
	UserPhoneNumber
)

// NewConnection attempt to create a database connection with the provided url
func NewConnection(url string) (*sqlx.DB, error) {
	if url == "" {
		return nil, ErrInvalidDSN
	}

	db, err := sqlx.Connect("postgres", string(url))
	if err != nil {
		return nil, fmt.Errorf("new connection: %w", err)
	}
	return db, nil
}

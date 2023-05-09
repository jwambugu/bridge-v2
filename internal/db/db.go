package db

import (
	"bridge/internal/config"
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

var (
	user     = config.Get[string](config.DbUser, "")
	password = config.Get[string](config.DbPassword, "")
	host     = config.Get[string](config.DbHost, "")
	name     = config.Get[string](config.DbName, "")

	url = config.Key(fmt.Sprintf(`postgres://%v:%v@%v/%v?sslmode=disable`, user, password, host, name))
)

// NewConnection attempt to create a database connection with the provided config.DbURL
func NewConnection() (*sqlx.DB, error) {
	if url == "" {
		return nil, ErrInvalidDSN
	}

	db, err := sqlx.Connect("postgres", string(url))
	if err != nil {
		return nil, fmt.Errorf("new connection: %w", err)
	}
	return db, nil
}

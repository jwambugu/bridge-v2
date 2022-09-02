package config

// Key represents an env key which should match the key on the .env file
type Key string

const (
	DbURL Key = "DB_URL"
)

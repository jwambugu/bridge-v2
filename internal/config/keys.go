package config

// Key represents an env key which should match the key on the .env file
type Key string

const (
	AppEnv  Key = "APP_ENV"
	AppName Key = "APP_NAME"
	AppURL  Key = "APP_URL"

	DbName     Key = "DB_NAME"
	DbUser     Key = "DB_USER"
	DbPassword Key = "DB_PASSWORD"
	DbHost     Key = "DB_HOST"

	GrpcGWPort Key = "GRPC_GW_PORT"
	GrpcPort   Key = "GRPC_PORT"

	JWTKey Key = "JWT_SYMMETRIC_KEY"
)

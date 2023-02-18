package config

// Key represents an env key which should match the key on the .env file
type Key string

const (
	AppEnv  Key = "APP_ENV"
	AppName Key = "APP_NAME"
	AppURL  Key = "APP_URL"

	DbURL Key = "DB_URL"

	GrpcGWPort Key = "GRPC_GW_PORT"
	GrpcPort   Key = "GRPC_PORT"

	JWTKey Key = "JWT_SYMMETRIC_KEY"
)

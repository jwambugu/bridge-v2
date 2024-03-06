package main

import (
	"bridge/internal/config"
	"bridge/internal/db"
	"bridge/internal/logger"
	"context"
	"flag"
	"github.com/pressly/goose"
	"os"
	"time"
)

var (
	flags = flag.NewFlagSet("goose", flag.ExitOnError)
)

const _migrateDown = "down"

func main() {
	appLogger := logger.NewLogger()

	if err := flags.Parse(os.Args[1:]); err != nil {
		appLogger.Fatal().Err(err).Msg("parse flags")
	}

	args := flags.Args()
	if len(args) < 1 {
		flags.Usage()
		return
	}

	var arguments []string
	if len(args) > 0 {
		arguments = append(arguments, args[1:]...)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := config.NewDefaultConfig(ctx)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("get default config")
	}

	dbConn, err := db.NewConnection(config.EnvKey.DbDsn)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("create db connection")
	}

	var (
		command = args[0]
		env     = config.EnvKey.Env
	)

	if command == _migrateDown && env == config.ProductionEnvironment {
		appLogger.Fatal().Err(err).Msgf("Command %q cannot be applied on environment %q", command, env)
	}

	appLogger.Info().Str("command", command).Msg("running command")

	if err = goose.Run(command, dbConn.DB, "./internal/db/migrations", arguments...); err != nil {
		appLogger.Fatal().Err(err).Msgf("goose run %v: %v", command, err)
	}
}

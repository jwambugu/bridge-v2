package main

import (
	"bridge/pkg/config"
	"bridge/pkg/db"
	"flag"
	"log"
	"os"

	"github.com/pressly/goose"
)

var (
	flags = flag.NewFlagSet("goose", flag.ExitOnError)
)

const _migrateDown = "down"

func main() {
	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatalf("parse flags: %v", err)
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

	dbConn, err := db.NewConnection()
	if err != nil {
		log.Fatalf("db: %v", err)
	}

	var (
		command = args[0]
		env     = config.GetEnvironment()
	)

	if command == _migrateDown && env == config.Production {
		log.Fatalf("Command %q cannot be applied on environment %q", command, env)
	}

	if err = goose.Run(command, dbConn.DB, "./pkg/db/migrations", arguments...); err != nil {
		log.Fatalf("goose run %v: %v", command, err)
	}
}

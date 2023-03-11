#!/bin/sh

set -e

echo "running db migrations"
./goose up

echo "starting the app.."
exec "$@"
#!/bin/bash
GOOSE=$(pwd)/bin/tools/goose

function status() {
  $GOOSE status
}

function create() {
    if [ -z "$*" ]; then
        echo "> Migration name is required"
    else
        $GOOSE create "$*" sql
    fi
}

function up() {
    $GOOSE up
}

function down() {
  $GOOSE down
}

$1 "$2"
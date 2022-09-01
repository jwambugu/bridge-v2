PWD := $(shell pwd)
APP_BIN := $(PWD)/bin
MIGRATION_TOOL := goose
MIGRATION_TOOL_FILES := $(PWD)/cmd/$(MIGRATION_TOOL)/main.go

.phony: build-goose
build-goose:
	@echo "  >  Building migration tool..."
	@go build -o $(APP_BIN)/tools/$(MIGRATION_TOOL) $(MIGRATION_TOOL_FILES)

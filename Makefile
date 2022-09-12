PWD := $(shell pwd)
APP_BIN := $(PWD)/bin
MIGRATION_TOOL := goose
MIGRATION_TOOL_FILES := $(PWD)/cmd/$(MIGRATION_TOOL)/main.go

.phony: build-goose
build-goose:
	@echo "  >  Building migration tool..."
	@go build -o $(APP_BIN)/tools/$(MIGRATION_TOOL) $(MIGRATION_TOOL_FILES)

.PHONY: protoc
protoc:
	rm -r api/v1/pb/*
	protoc -I api/v1 api/v1/*.proto  --go_out=api/v1 --go-grpc_out=api/v1 --validate_out="lang=go:./api/v1"
	cd api/v1/pb; protoc-go-inject-tag -input="*.pb.go" -remove_tag_comment

.PHONY: audit
audit:
	go install github.com/favadi/protoc-go-inject-tag@latest; which protoc-go-inject-tag
	go install github.com/envoyproxy/protoc-gen-validate@latest; which protoc-gen-validate;

.PHONY: test
test:
	 go clean -testcache && go test -cover -race ./...
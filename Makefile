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
	protoc -I api/v1 api/v1/*.proto --go_out=api/v1 --go-grpc_out=api/v1 \
 	--validate_out="lang=go:./api/v1" --grpc-gateway_out=api/v1
	cd api/v1/pb; protoc-go-inject-tag -input="*.pb.go" -remove_tag_comment

.PHONY: audit
audit:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28; which protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2; which protoc-gen-go-grpc
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest; which protoc-gen-grpc-gateway
	go install github.com/favadi/protoc-go-inject-tag@latest; which protoc-go-inject-tag
	go install github.com/envoyproxy/protoc-gen-validate@latest; which protoc-gen-validate;

.PHONY: test
test:
	go clean -testcache && go test -cover -race ./...

.PHONY: docker-build
docker-build:
	docker build -t bridge --no-cache .

.PHONY: compose-up
compose-up:
	docker-compose up -d --build
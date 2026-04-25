APP_NAME := mcp-server
BIN_DIR := bin

.PHONY: fmt vet test lint build run tidy

fmt:
	gofmt -w $$(go list -f '{{.Dir}}' ./...)

vet:
	go vet ./...

test:
	go test ./...

lint: fmt vet test

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/mcp-server

run:
	go run ./cmd/mcp-server

tidy:
	go mod tidy

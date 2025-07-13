# DevHub Makefile

.PHONY: build test clean install dev lint fmt vet deps

# Build variables
BINARY_NAME=devhub
VERSION=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse HEAD)
DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
all: test build

# Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/devhub

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/devhub
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/devhub
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/devhub
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/devhub

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install the binary
install: build
	go install $(LDFLAGS) ./cmd/devhub

# Development mode (with live reload)
dev:
	go run $(LDFLAGS) ./cmd/devhub

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	go fmt ./...

# Vet the code
vet:
	go vet ./...

# Security check
sec:
	gosec ./...

# Run all checks
check: fmt vet lint test

# Build Docker image
docker:
	docker build -t devhub:$(VERSION) .


# Development dependencies
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

# Show help
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  build-all   - Build for multiple platforms"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean build artifacts"
	@echo "  install     - Install the binary"
	@echo "  dev         - Run in development mode"
	@echo "  lint        - Lint the code"
	@echo "  fmt         - Format the code"
	@echo "  vet         - Vet the code"
	@echo "  check       - Run all checks"
	@echo "  deps        - Install dependencies"
	@echo "  dev-deps    - Install development dependencies"
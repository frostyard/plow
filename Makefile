.PHONY: all build test lint clean install

# Build variables
BINARY_NAME := plow
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/frostyard/plow/internal/cli.Version=$(VERSION) \
	-X github.com/frostyard/plow/internal/cli.Commit=$(COMMIT) \
	-X github.com/frostyard/plow/internal/cli.BuildDate=$(BUILD_DATE)

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOVET := $(GOCMD) vet
GOMOD := $(GOCMD) mod
GOFMT := gofmt

# Default target
all: lint test build

# Build the binary
build:
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/plow

# Run tests
test:
	$(GOTEST) -v -race -cover ./...

# Run tests with coverage report
test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run linters
lint: vet fmt-check
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed, skipping (install: https://golangci-lint.run/usage/install/)"; \
	fi

# Run go vet
vet:
	$(GOVET) ./...

# Check formatting
fmt-check:
	@if [ -n "$$($(GOFMT) -l .)" ]; then \
		echo "The following files need formatting:"; \
		$(GOFMT) -l .; \
		exit 1; \
	fi

# Format code
fmt:
	$(GOFMT) -w .

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Install binary to $GOPATH/bin
install:
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(GOPATH)/bin/$(BINARY_NAME) ./cmd/plow

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)_linux_amd64 ./cmd/plow
	GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)_linux_arm64 ./cmd/plow
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)_darwin_amd64 ./cmd/plow
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)_darwin_arm64 ./cmd/plow

# Help
help:
	@echo "Available targets:"
	@echo "  all           - Run lint, test, and build (default)"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests with race detection"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run all linters (vet, fmt-check, golangci-lint)"
	@echo "  vet           - Run go vet"
	@echo "  fmt           - Format code"
	@echo "  fmt-check     - Check code formatting"
	@echo "  tidy          - Tidy go.mod"
	@echo "  clean         - Remove build artifacts"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  build-all     - Build for all platforms"
	@echo "  help          - Show this help"

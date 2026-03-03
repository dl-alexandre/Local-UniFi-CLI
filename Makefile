.PHONY: build build-all build-linux build-darwin build-windows test lint release clean format install-hooks

BINARY_NAME=unifi
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME) -s -w"

# Build for current platform
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/unifi

# Build for all platforms
build-all: build-linux build-darwin build-windows

# Linux builds
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/unifi
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/unifi

# macOS builds
build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/unifi
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/unifi

# Windows builds
build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/unifi

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run integration tests (requires API access)
test-integration:
	go test -v -tags=integration ./...

# Run linter
lint:
	golangci-lint run ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f coverage.out

# Release build (optimized)
release: clean
	CGO_ENABLED=0 go build $(LDFLAGS) -trimpath -o $(BINARY_NAME) ./cmd/unifi

# Development build with debug info
dev:
	go build -o $(BINARY_NAME) ./cmd/unifi

# Install locally
install: build
	go install ./...

# Format code
format:
	@echo "Formatting code..."
	@gofmt -w -s .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "goimports not installed. Install: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

# Install git hooks
install-hooks:
	@echo "Installing git hooks..."
	@git config core.hooksPath .githooks
	@echo "Hooks installed from .githooks/"

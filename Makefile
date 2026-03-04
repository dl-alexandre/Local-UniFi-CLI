# Variables
BINARY_NAME=unifi
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT} -s -w"

# Default target
.PHONY: all
all: test build

# Build binary
.PHONY: build
build:
	@echo "Building ${BINARY_NAME}..."
	go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/${BINARY_NAME}

# Build for all platforms
.PHONY: build-all
build-all: build-linux build-darwin build-windows

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 ./cmd/${BINARY_NAME}
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-arm64 ./cmd/${BINARY_NAME}

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-amd64 ./cmd/${BINARY_NAME}
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-arm64 ./cmd/${BINARY_NAME}

.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-windows-amd64.exe ./cmd/${BINARY_NAME}

# Testing
.PHONY: test
test:
	@echo "Running tests..."
	go test -v -race ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: test-short
test-short:
	@echo "Running short tests..."
	go test -short ./...

# Linting
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/"; \
	fi

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Clean
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf bin/ dist/ coverage.out coverage.html
	go clean

# Install locally
.PHONY: install
install: build
	@echo "Installing to /usr/local/bin..."
	sudo cp bin/${BINARY_NAME} /usr/local/bin/

.PHONY: uninstall
uninstall:
	@echo "Uninstalling..."
	sudo rm -f /usr/local/bin/${BINARY_NAME}

# Docker
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t ${BINARY_NAME}:${VERSION} .

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run --rm -it ${BINARY_NAME}:${VERSION}

.PHONY: docker-compose-up
docker-compose-up:
	@echo "Starting with docker-compose..."
	docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down:
	@echo "Stopping docker-compose..."
	docker-compose down

# Development
.PHONY: dev
dev:
	@echo "Running in development mode..."
	go run ./cmd/${BINARY_NAME}

.PHONY: run
run: build
	@echo "Running binary..."
	./bin/${BINARY_NAME}

# Release
.PHONY: release
release: clean test build-all
	@echo "Creating release artifacts..."
	mkdir -p dist
	for file in bin/*; do \
		if [[ "$$file" == *windows* ]]; then \
			zip "dist/$$(basename $$file).zip" "$$file"; \
		else \
			tar -czf "dist/$$(basename $$file).tar.gz" -C bin "$$(basename $$file)"; \
		fi; \
	done
	@echo "Release artifacts in dist/"

# Security
.PHONY: security
security:
	@echo "Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Run: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

.PHONY: vulncheck
vulncheck:
	@echo "Checking for vulnerabilities..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed. Run: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# Documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Starting godoc server on http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "godoc not installed. Run: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make build          - Build binary for current platform"
	@echo "  make build-all      - Build for all platforms (Linux, macOS, Windows)"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run linter"
	@echo "  make fmt            - Format code"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make install        - Install binary to /usr/local/bin"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make release        - Create release artifacts"
	@echo "  make help           - Show this help"

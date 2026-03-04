#!/bin/bash
set -e

echo "Setting up development environment..."

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "✅ Go version: $GO_VERSION"

# Install development tools
echo "Installing development tools..."

# golangci-lint
if ! command -v golangci-lint &> /dev/null; then
    echo "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

# gosec
if ! command -v gosec &> /dev/null; then
    echo "Installing gosec..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

# govulncheck
if ! command -v govulncheck &> /dev/null; then
    echo "Installing govulncheck..."
    go install golang.org/x/vuln/cmd/govulncheck@latest
fi

# Download dependencies
echo "Downloading Go dependencies..."
go mod download
go mod tidy

# Create directories
mkdir -p bin dist

echo "✅ Development environment ready!"
echo ""
echo "Next steps:"
echo "  1. Copy .env.example to .env and configure"
echo "  2. Run 'make build' to build"
echo "  3. Run 'make test' to test"
echo "  4. Run 'make help' to see all commands"

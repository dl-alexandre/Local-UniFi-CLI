# Contributing to Local UniFi CLI

Thank you for considering contributing! This document provides guidelines and information to help you get started.

## Getting Started

### Prerequisites

- Go 1.24 or later
- Docker (optional, for containerized development)
- Access to a UniFi Controller (Cloud Key, Dream Machine, or Software Controller)

### Setting Up Your Development Environment

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/Local-UniFi-CLI.git
   cd Local-UniFi-CLI
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Build the project:
   ```bash
   go build -o unifi ./cmd/unifi
   ```

5. Run tests:
   ```bash
   go test ./...
   ```

## Project Structure

```
Local-UniFi-CLI/
├── cmd/unifi/           # Main application entry point
├── internal/
│   ├── cli/             # CLI command implementations
│   ├── api/             # UniFi API client
│   ├── config/          # Configuration management
│   └── output/          # Output formatting utilities
├── completions/          # Shell completion scripts
├── config.example.yaml   # Example configuration
├── Dockerfile           # Docker build configuration
├── docker-compose.yml   # Docker Compose for testing
└── Makefile             # Build automation
```

## Development Workflow

### Building

```bash
# Build for current platform
go build -o unifi ./cmd/unifi

# Build all platforms (requires GoReleaser)
make build-all

# Build with Docker
docker-compose build
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...

# Run specific test
go test -v ./internal/api/...

# Run linter
make lint

# Run security scan
make security
```

### Configuration for Testing

Create a `.env` file or use environment variables:

```bash
export UNIFI_BASE_URL="https://unifi.local:8443"
export UNIFI_USERNAME="admin"
export UNIFI_PASSWORD="your-password"
export UNIFI_SITE="default"
```

## Code Style

- Follow standard Go conventions (`gofmt`, `golint`)
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and small
- Use table-driven tests where appropriate

## Making Changes

1. Create a new branch for your feature or bug fix:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes and ensure tests pass

3. Commit your changes with a descriptive message:
   ```bash
   git commit -m "feat: add support for XYZ"
   ```

4. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

5. Open a Pull Request against the main repository

## Commit Message Convention

We use conventional commits. Format: `<type>(<scope>): <description>`

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Test additions or modifications
- `chore`: Build process or auxiliary tool changes

Examples:
- `feat: add device restart command`
- `fix: handle empty site list`
- `docs: update README with Docker instructions`

## Pull Request Guidelines

- Ensure all tests pass
- Update documentation if needed
- Add tests for new functionality
- Follow the existing code style
- Keep PRs focused on a single concern
- Reference any related issues

## Reporting Bugs

When reporting bugs, please include:

- CLI version (`unifi --version`)
- Go version (`go version`)
- Operating system and version
- UniFi Controller version
- Steps to reproduce
- Expected behavior
- Actual behavior
- Debug output (run with `--debug` flag, redact sensitive info)

## Requesting Features

When requesting features, please:

- Describe the use case
- Explain why current functionality is insufficient
- Provide examples of how the feature would work
- Consider implementation complexity

## Security

- Never commit credentials or API keys
- Use environment variables or config files for sensitive data
- Report security vulnerabilities privately to the maintainers

## Questions?

Feel free to open an issue with the `question` label if you need help.

## Release Checklist

Before a new release:
- [ ] All tests pass (`go test ./...`)
- [ ] Build succeeds for all platforms
- [ ] Version bumped in `cmd/unifi/main.go`
- [ ] CHANGELOG.md updated (if exists)
- [ ] Documentation updated
- [ ] Tag created: `git tag -a vX.X.X -m "Release vX.X.X"`
- [ ] Tag pushed: `git push origin vX.X.X`

---

Thank you for contributing to Local UniFi CLI!

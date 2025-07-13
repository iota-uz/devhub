# Contributing to DevHub

Thank you for your interest in contributing to DevHub! This document provides guidelines and information for contributors.

## Development Setup

### Prerequisites

- Go 1.23 or later
- Git

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/iota-uz/devhub.git
   cd devhub
   ```

2. Install dependencies:
   ```bash
   make deps
   ```

3. Install development tools:
   ```bash
   make dev-deps
   ```

4. Run in development mode:
   ```bash
   make dev
   ```

## Development Workflow

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting: `make fmt`
- Run linters: `make lint`
- Run static analysis: `make vet`

### Testing

- Write tests for new functionality
- Run tests: `make test`
- Check test coverage: `make test-coverage`

### Before Submitting

Run all checks to ensure code quality:
```bash
make check
```

This runs formatting, vetting, linting, and tests.

## Installation

Users can install DevHub via:

**Go install** (recommended):
```bash
go install github.com/iota-uz/devhub/cmd/devhub@latest
```

**Build from source**:
```bash
git clone https://github.com/iota-uz/devhub.git
cd devhub
make build
```

## Project Structure

See [CLAUDE.md](CLAUDE.md) for detailed architecture documentation including:

- Core components and responsibilities
- Service implementation patterns
- MCP integration details
- Configuration format
- Testing guidelines

## Pull Request Process

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes following the development guidelines
4. Run `make check` to ensure code quality
5. Commit your changes with descriptive messages
6. Push to your fork and create a pull request
7. Address any feedback from code review

## Issues and Feature Requests

- Check existing issues before creating new ones
- Use issue templates when available
- Provide clear reproduction steps for bugs
- Include relevant system information (OS, Go version, etc.)

## Code of Conduct

Please be respectful and constructive in all interactions. We want to maintain a welcoming environment for all contributors.

## Questions?

Feel free to open an issue for questions about contributing or join discussions in existing issues.
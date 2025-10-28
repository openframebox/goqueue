# Contributing to GoQueue

Thank you for your interest in contributing to GoQueue! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Clear title and description**
- **Steps to reproduce** the issue
- **Expected behavior** vs **actual behavior**
- **Code samples** if applicable
- **Environment details** (Go version, OS, backend used)

### Suggesting Enhancements

Enhancement suggestions are welcome! Please provide:

- **Clear use case** for the feature
- **Detailed description** of the proposed functionality
- **Potential implementation approach** (if you have ideas)
- **Why this would be useful** to other users

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following our coding standards
3. **Add tests** for any new functionality
4. **Update documentation** if needed
5. **Ensure tests pass** (`go test ./...`)
6. **Submit your pull request**

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Redis (for testing Redis backend)
- Git

### Getting Started

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/goqueue.git
cd goqueue

# Install dependencies
go mod download

# Run tests
go test ./...

# Run examples
go run examples/basic/main.go
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detector
go test -race ./...

# Run specific test
go test -run TestFunctionName ./...
```

### Testing with Redis

For Redis backend tests:

```bash
# Start Redis (if not running)
redis-server

# Run Redis tests
go test -v ./... -tags=redis
```

## Coding Standards

### Code Style

- Follow standard Go conventions and `gofmt` formatting
- Use `golangci-lint` for linting
- Write clear, self-documenting code
- Add comments for exported functions and types
- Keep functions focused and concise

### Naming Conventions

- Use descriptive names for variables and functions
- Follow Go naming conventions (camelCase for unexported, PascalCase for exported)
- Interfaces should describe behavior (e.g., `Handler`, `Backend`)

### Error Handling

- Return errors instead of panicking
- Provide context in error messages
- Use `fmt.Errorf` with `%w` for error wrapping

Example:
```go
if err != nil {
    return fmt.Errorf("failed to publish message: %w", err)
}
```

### Testing

- Write table-driven tests where appropriate
- Test both success and failure cases
- Use meaningful test names
- Mock external dependencies

Example test structure:
```go
func TestPublish(t *testing.T) {
    tests := []struct {
        name    string
        message QueueMessage
        want    error
    }{
        {
            name:    "valid message",
            message: &TestMessage{},
            want:    nil,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Creating a New Backend

To add support for a new queue backend:

1. **Create a new file** (e.g., `rabbitmq.go`)
2. **Implement the `Backend` interface**:
   ```go
   type Backend interface {
       Publish(ctx context.Context, queue string, envelope *Envelope) error
       Subscribe(ctx context.Context, queue string) (<-chan *Envelope, error)
       Ack(ctx context.Context, messageID string) error
       Nack(ctx context.Context, messageID string) error
       Close() error
   }
   ```
3. **Add configuration options** using the functional options pattern
4. **Write comprehensive tests**
5. **Add example usage** in `examples/` directory
6. **Update documentation** in README.md

## Documentation

- Update README.md for user-facing changes
- Add inline documentation for exported functions
- Include code examples where helpful
- Update CHANGELOG.md with your changes

## Commit Messages

Write clear, meaningful commit messages:

```
Add support for RabbitMQ backend

- Implement Backend interface for RabbitMQ
- Add connection pooling and reconnection logic
- Include example usage
- Add unit tests with 90% coverage

Closes #123
```

Format:
- **First line**: Brief summary (50 chars or less)
- **Body**: Detailed explanation (wrap at 72 chars)
- **Footer**: Reference issues/PRs

## Pull Request Process

1. **Update documentation** for any user-facing changes
2. **Add/update tests** to maintain coverage
3. **Ensure CI passes** before requesting review
4. **Link related issues** in PR description
5. **Respond to review comments** promptly
6. **Squash commits** if requested

### PR Checklist

Before submitting:

- [ ] Code follows project style guidelines
- [ ] Tests added/updated and passing
- [ ] Documentation updated
- [ ] No breaking changes (or documented if unavoidable)
- [ ] Commit messages are clear
- [ ] Branch is up to date with main

## Release Process

Releases are managed by maintainers:

1. Version bump in accordance with [Semantic Versioning](https://semver.org/)
2. Update CHANGELOG.md
3. Create GitHub release with release notes
4. Tag release in Git

## Getting Help

- **Questions?** Open a discussion on GitHub
- **Bugs?** Create an issue with details
- **Ideas?** Start a discussion or create an enhancement issue

## Recognition

Contributors will be:
- Listed in the project's contributor list
- Mentioned in release notes for significant contributions
- Credited in the CHANGELOG

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to GoQueue! 🎉

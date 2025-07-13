# Contributing to KubeJanitor

Thank you for your interest in contributing to KubeJanitor! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Process](#contributing-process)
- [Coding Guidelines](#coding-guidelines)
- [Testing](#testing)
- [Documentation](#documentation)
- [Release Process](#release-process)

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). Please read it before contributing.

## Getting Started

### Ways to Contribute

- 🐛 **Bug Reports**: Report bugs using GitHub issues
- 💡 **Feature Requests**: Suggest new features or improvements
- 📖 **Documentation**: Improve documentation, examples, or tutorials
- 🔧 **Code Contributions**: Fix bugs, implement features, or improve performance
- 🧪 **Testing**: Add test cases or improve test coverage
- 🎨 **Design**: Improve UI/UX for dashboards or web interfaces

### Before You Start

1. Check [existing issues](https://github.com/automationpi/kubejanitor/issues) to avoid duplicates
2. For major changes, create an issue first to discuss the approach
3. Fork the repository and create a feature branch
4. Ensure you have the necessary development environment set up

## Development Setup

### Prerequisites

- Go 1.21+
- Docker
- kubectl
- kind (for local testing)
- Helm 3.x
- make

### Local Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/automationpi/kubejanitor.git
   cd kubejanitor
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Set up development tools**:
   ```bash
   make install-tools
   ```

4. **Create a local test cluster**:
   ```bash
   make test-cluster-up
   ```

5. **Build and deploy locally**:
   ```bash
   make build
   make deploy-local
   ```

### Development Workflow

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**:
   - Write code following our [coding guidelines](#coding-guidelines)
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**:
   ```bash
   make test
   make test-e2e
   ```

4. **Lint and format**:
   ```bash
   make lint
   make fmt
   ```

5. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add new cleanup feature"
   ```

6. **Push and create a PR**:
   ```bash
   git push origin feature/your-feature-name
   ```

## Contributing Process

### Issue Guidelines

When creating an issue, please:

1. **Use a clear, descriptive title**
2. **Provide detailed description** including:
   - Steps to reproduce (for bugs)
   - Expected vs actual behavior
   - Environment details (Kubernetes version, operator version)
   - Relevant logs or error messages

3. **Use issue templates** when available
4. **Add appropriate labels** (bug, enhancement, documentation, etc.)

### Pull Request Guidelines

1. **Link to related issues** using "Fixes #123" or "Addresses #123"
2. **Provide clear description** of changes
3. **Include tests** for new functionality
4. **Update documentation** if needed
5. **Ensure CI passes** before requesting review
6. **Keep PRs focused** - one feature/fix per PR
7. **Respond to feedback** promptly

### PR Title Format

Use conventional commit format:
- `feat: add new PVC cleanup feature`
- `fix: resolve crash loop detection issue`
- `docs: update installation guide`
- `test: add integration tests for jobs cleanup`
- `refactor: improve error handling`

## Coding Guidelines

### Go Style

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Run `golangci-lint` and fix all issues
- Write clear, self-documenting code

### Code Organization

```
├── api/                    # CRD definitions
├── controllers/            # Controller logic
├── pkg/                    # Shared packages
│   ├── cleanup/           # Cleanup engines
│   ├── metrics/           # Metrics collection
│   └── utils/             # Utility functions
├── config/                 # Kubernetes manifests
├── helm-charts/           # Helm charts
├── test/                  # Test files
├── docs/                  # Documentation
└── scripts/               # Build and utility scripts
```

### Error Handling

- Use structured errors with context
- Log errors at appropriate levels
- Return meaningful error messages
- Handle edge cases gracefully

Example:
```go
if err := r.Client.Get(ctx, req.NamespacedName, &policy); err != nil {
    if errors.IsNotFound(err) {
        log.Info("JanitorPolicy not found, ignoring")
        return ctrl.Result{}, nil
    }
    log.Error(err, "Failed to get JanitorPolicy", "policy", req.NamespacedName)
    return ctrl.Result{}, err
}
```

### Logging

- Use structured logging with contextual fields
- Include relevant metadata (resource names, namespaces)
- Use appropriate log levels

Example:
```go
log.Info("Starting cleanup",
    "policy", policy.Name,
    "namespace", policy.Namespace,
    "dryRun", policy.Spec.DryRun)
```

### Documentation

- Add GoDoc comments for exported functions and types
- Include examples in complex functions
- Document configuration options
- Update README and docs for user-facing changes

## Testing

### Test Types

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test complete workflows

### Running Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests
make test-integration

# Run e2e tests
make test-e2e

# Run tests with coverage
make test-coverage
```

### Test Guidelines

- Write tests for all new functionality
- Aim for >80% code coverage
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Test error conditions

Example:
```go
func TestPVCCleaner_Execute(t *testing.T) {
    tests := []struct {
        name     string
        setup    func(*fake.Client)
        policy   *v1alpha1.JanitorPolicy
        expected *v1alpha1.ResourceTypeStats
        wantErr  bool
    }{
        {
            name: "cleanup unused PVC",
            setup: func(client *fake.Client) {
                // Setup test data
            },
            policy: &v1alpha1.JanitorPolicy{
                // Test policy
            },
            expected: &v1alpha1.ResourceTypeStats{
                Scanned: 1,
                Cleaned: 1,
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Documentation

### Types of Documentation

1. **Code Documentation**: GoDoc comments
2. **User Documentation**: Installation, configuration, examples
3. **Developer Documentation**: Architecture, development setup
4. **API Documentation**: CRD reference

### Documentation Standards

- Use clear, concise language
- Include practical examples
- Keep documentation up-to-date with code changes
- Use proper markdown formatting
- Include diagrams where helpful

### Building Documentation

```bash
# Generate API documentation
make generate-docs

# Serve documentation locally
make serve-docs
```

## Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Checklist

1. Update version in relevant files
2. Update CHANGELOG.md
3. Create release PR
4. Tag release after PR merge
5. GitHub Actions handles the rest

### Creating a Release

1. **Update version**:
   ```bash
   # Update Chart.yaml, go.mod, etc.
   ```

2. **Update changelog**:
   ```bash
   # Add release notes to CHANGELOG.md
   ```

3. **Create release PR**:
   ```bash
   git checkout -b release/v1.2.3
   git commit -m "chore: prepare release v1.2.3"
   git push origin release/v1.2.3
   ```

4. **Tag release**:
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```

## Getting Help

- 💬 [GitHub Discussions](https://github.com/automationpi/kubejanitor/discussions) for questions
- 🐛 [GitHub Issues](https://github.com/automationpi/kubejanitor/issues) for bugs
- 📧 Email: [hello@automationpi.com](mailto:hello@automationpi.com)

## Recognition

Contributors will be recognized in:
- CONTRIBUTORS.md file
- Release notes
- GitHub contributors page

Thank you for contributing to KubeJanitor! 🎉
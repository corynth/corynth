# Contributing to Corynth

Thank you for considering contributing to Corynth! This document provides guidelines and information for contributors.

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## How to Contribute

### Reporting Bugs

1. **Search existing issues** first to avoid duplicates
2. **Use the issue template** when creating new bug reports
3. **Include reproduction steps** and environment details
4. **Provide logs and error messages** when possible

### Suggesting Features

1. **Check the roadmap** to see if the feature is already planned
2. **Create a detailed proposal** with use cases and examples
3. **Discuss the feature** in issues before implementation
4. **Consider backward compatibility** and breaking changes

### Code Contributions

#### Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/corynth.git
   cd corynth
   ```
3. **Create a feature branch**:
   ```bash
   git checkout -b feature/amazing-feature
   ```

#### Development Setup

1. **Install Go 1.21+**
2. **Install dependencies**:
   ```bash
   go mod download
   ```
3. **Build the project**:
   ```bash
   make build
   ```
4. **Run tests**:
   ```bash
   make test
   ```

#### Code Standards

- **Follow Go conventions** and idioms
- **Add tests** for new functionality
- **Update documentation** for user-facing changes  
- **Use conventional commits** for commit messages
- **Keep commits focused** - one logical change per commit

#### Commit Message Format

```
type(scope): description

Detailed explanation of the change (if needed)

Fixes #123
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

#### Testing

- **Unit tests**: `go test ./...`
- **Integration tests**: `make test-integration`
- **Plugin tests**: Test custom plugins with sample workflows
- **Cross-platform**: Verify changes work on Linux, macOS, Windows

#### Pull Request Process

1. **Update documentation** if needed
2. **Add tests** for new functionality
3. **Ensure all tests pass**
4. **Update CHANGELOG.md** if applicable
5. **Create pull request** with clear description
6. **Respond to review feedback** promptly

### Plugin Development

#### Creating Plugins

1. **Follow the plugin development guide** in `/docs/plugins/development-guide.md`
2. **Use the plugin generator**:
   ```bash
   corynth plugin init my-plugin --type http
   ```
3. **Test thoroughly** before submitting
4. **Document all actions** and parameters

#### Plugin Guidelines

- **Implement all required interfaces**
- **Handle errors gracefully**
- **Validate input parameters**
- **Include comprehensive tests**
- **Provide clear documentation**
- **Follow security best practices**

## Development Guidelines

### Architecture Principles

- **Plugin-based architecture** for extensibility
- **HCL configuration** for declarative workflows
- **State management** for execution tracking
- **Security first** approach
- **Performance** and efficiency focus

### Code Organization

```
corynth/
├── cmd/corynth/          # Main application entry point
├── pkg/                  # Core packages
│   ├── cli/             # Command line interface
│   ├── config/          # Configuration management
│   ├── plugin/          # Plugin system
│   ├── state/           # State management
│   └── workflow/        # Workflow engine
├── docs/                # Documentation
├── examples/            # Example workflows
├── test/               # Test files and test utilities
└── scripts/            # Build and utility scripts
```

### Testing Strategy

- **Unit tests** for individual components
- **Integration tests** for plugin interactions
- **End-to-end tests** for complete workflows
- **Performance tests** for scalability
- **Security tests** for vulnerability assessment

## Documentation

### Types of Documentation

- **User documentation**: Installation, usage, tutorials
- **Developer documentation**: APIs, architecture, contributing
- **Plugin documentation**: Development guides, references
- **Example workflows**: Real-world use cases

### Documentation Standards

- **Clear and concise** writing
- **Code examples** for all features
- **Up-to-date** with current functionality
- **Accessible** to different skill levels
- **Well-organized** and searchable

## Community

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General discussion and questions
- **Documentation**: Comprehensive guides and references

### Getting Help

1. **Check documentation** first
2. **Search existing issues** and discussions
3. **Create new issue** with detailed information
4. **Be patient and respectful** in communications

## Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):
- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Checklist

- [ ] Update version numbers
- [ ] Update CHANGELOG.md
- [ ] Create release notes
- [ ] Build cross-platform binaries
- [ ] Test release candidates
- [ ] Create GitHub release
- [ ] Update documentation

## Legal

### License

By contributing to Corynth, you agree that your contributions will be licensed under the Apache License 2.0.

### Contributor License Agreement

For significant contributions, we may require a Contributor License Agreement (CLA) to ensure proper licensing.

## Questions?

If you have questions about contributing, please:

1. Check this document first
2. Search existing issues and discussions
3. Create a new discussion or issue
4. Reach out to maintainers if needed

Thank you for contributing to Corynth! 🚀
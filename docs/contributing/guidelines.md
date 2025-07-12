# Contributing Guidelines

Thank you for your interest in contributing to Tilt-Valid! This document provides guidelines for contributing to the project.

## 🤝 How to Contribute

### Types of Contributions

We welcome the following types of contributions:

- **🐛 Bug Reports**: Report bugs and issues
- **✨ Feature Requests**: Suggest new features and improvements
- **📝 Documentation**: Improve or add documentation
- **🔧 Code Contributions**: Submit code changes and improvements
- **🧪 Testing**: Add tests or improve test coverage
- **📊 Performance**: Optimize performance and efficiency

### Before You Start

1. **Check Existing Issues**: Search existing issues to avoid duplicates
2. **Read Documentation**: Familiarize yourself with the project structure
3. **Set Up Development Environment**: Follow the [Development Setup Guide](../development/setup.md)
4. **Understand the Codebase**: Review the [Code Structure Guide](../development/code-structure.md)

## 🚀 Getting Started

### 1. Fork and Clone

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/your-username/tilt-validator.git
cd tilt-validator

# Add the original repository as upstream
git remote add upstream https://github.com/original-org/tilt-validator.git
```

### 2. Create a Branch

```bash
# Create a new branch for your feature
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/your-bug-description
```

### 3. Make Your Changes

- Follow the [Code Style Guide](./code-style.md)
- Write tests for new functionality
- Update documentation as needed
- Ensure all tests pass

### 4. Test Your Changes

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Test the complete flow
./cmd/run_validators.sh
```

### 5. Commit Your Changes

```bash
# Add your changes
git add .

# Commit with a descriptive message
git commit -m "feat: add new tilt type for complex distributions

- Add support for multi-level nested tilts
- Implement recursive distribution algorithm
- Add comprehensive test coverage
- Update documentation with examples"
```

### 6. Push and Create Pull Request

```bash
# Push your branch
git push origin feature/your-feature-name

# Create a pull request on GitHub
```

## 📋 Pull Request Guidelines

### Before Submitting

- [ ] **Tests Pass**: All tests should pass
- [ ] **Code Style**: Follow the project's code style
- [ ] **Documentation**: Update relevant documentation
- [ ] **No Breaking Changes**: Ensure backward compatibility
- [ ] **Security**: No security vulnerabilities introduced

### Pull Request Template

```markdown
## Description

Brief description of the changes made.

## Type of Change

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing

- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed
- [ ] Performance impact assessed

## Checklist

- [ ] My code follows the style guidelines of this project
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published

## Additional Notes

Any additional information or context for reviewers.
```

## 🐛 Bug Reports

### Before Reporting

1. **Check Existing Issues**: Search for similar issues
2. **Reproduce the Issue**: Ensure you can consistently reproduce it
3. **Gather Information**: Collect relevant logs and system information

### Bug Report Template

```markdown
## Bug Description

Clear and concise description of the bug.

## Steps to Reproduce

1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

## Expected Behavior

What you expected to happen.

## Actual Behavior

What actually happened.

## Environment

- OS: [e.g., macOS 12.0]
- Go Version: [e.g., 1.22.0]
- Tilt-Valid Version: [e.g., 1.0.0]

## Additional Context

Any other context about the problem, including logs, screenshots, etc.
```

## ✨ Feature Requests

### Feature Request Template

```markdown
## Feature Description

Clear and concise description of the feature.

## Problem Statement

What problem does this feature solve?

## Proposed Solution

How would you like to see this implemented?

## Alternatives Considered

Any alternative solutions or features you've considered.

## Additional Context

Any other context or screenshots about the feature request.
```

## 🧪 Testing Guidelines

### Writing Tests

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test complete workflows

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    // Arrange
    input := "test input"
    expected := "expected output"

    // Act
    result := FunctionName(input)

    // Assert
    if result != expected {
        t.Errorf("Expected %s, got %s", expected, result)
    }
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestFunctionName

# Run benchmarks
go test -bench=. ./...
```

## 📝 Documentation Guidelines

### When to Update Documentation

- **New Features**: Document new functionality
- **API Changes**: Update API documentation
- **Configuration Changes**: Update configuration guides
- **Bug Fixes**: Update troubleshooting guides if relevant

### Documentation Standards

1. **Clear and Concise**: Write in simple, clear language
2. **Examples**: Include practical examples
3. **Code Blocks**: Use proper syntax highlighting
4. **Links**: Link to related documentation
5. **Images**: Include diagrams when helpful

## 🔧 Code Style Guidelines

### Go Code Style

1. **Formatting**: Use `go fmt` for formatting
2. **Imports**: Use `goimports` for import organization
3. **Naming**: Follow Go naming conventions
4. **Comments**: Add comments for exported functions
5. **Error Handling**: Handle errors explicitly

### Example Code Style

```go
// Package mpc provides multi-party computation functionality.
package mpc

import (
    "context"
    "fmt"
)

// Party represents a participant in the MPC protocol.
type Party struct {
    ID   uint16
    Name string
}

// NewParty creates a new party with the given ID and name.
func NewParty(id uint16, name string) *Party {
    return &Party{
        ID:   id,
        Name: name,
    }
}

// Start begins the party's participation in the protocol.
func (p *Party) Start(ctx context.Context) error {
    if p.ID == 0 {
        return fmt.Errorf("invalid party ID: %d", p.ID)
    }

    // Implementation here
    return nil
}
```

## 🚀 Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

### Release Checklist

- [ ] All tests pass
- [ ] Documentation is updated
- [ ] Changelog is updated
- [ ] Version is bumped
- [ ] Release notes are written
- [ ] Release is tagged

## 🤝 Community Guidelines

### Communication

- **Be Respectful**: Treat others with respect and kindness
- **Be Constructive**: Provide constructive feedback
- **Be Patient**: Allow time for responses and reviews
- **Be Helpful**: Help others when you can

### Code of Conduct

- **Inclusive**: Welcome contributors from all backgrounds
- **Professional**: Maintain professional behavior
- **Constructive**: Focus on constructive feedback
- **Respectful**: Respect different opinions and approaches

## 📞 Getting Help

### Resources

- **Documentation**: Check the [docs](../) directory
- **Issues**: Search existing issues on GitHub
- **Discussions**: Use GitHub Discussions for questions
- **Code Review**: Ask for code review on pull requests

### Contact

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion
- **Email**: For sensitive or private matters

## 🎯 Contribution Areas

### High Priority

- **Performance Optimization**: Improve system performance
- **Security Enhancements**: Strengthen security measures
- **Test Coverage**: Increase test coverage
- **Documentation**: Improve and expand documentation

### Medium Priority

- **New Features**: Add new functionality
- **Bug Fixes**: Fix existing issues
- **Code Refactoring**: Improve code quality
- **Tooling**: Add development tools

### Low Priority

- **Cosmetic Changes**: UI/UX improvements
- **Documentation**: Minor documentation updates
- **Examples**: Add more examples
- **Tutorials**: Create tutorials

## 🙏 Recognition

### Contributors

We recognize all contributors in our:

- **README**: List of major contributors
- **Changelog**: Credit for contributions
- **Release Notes**: Acknowledge contributions
- **Contributors File**: Complete list of contributors

### Hall of Fame

Special recognition for:

- **Major Contributors**: Significant code contributions
- **Documentation Heroes**: Outstanding documentation work
- **Bug Hunters**: Finding and fixing critical bugs
- **Community Leaders**: Helping others contribute

---

**Thank you for contributing to Tilt-Valid!** 🚀

Your contributions help make this project better for everyone. We appreciate your time and effort.

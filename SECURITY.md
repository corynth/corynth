# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |

## Reporting a Vulnerability

The Corynth team takes security vulnerabilities seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report

Please report security vulnerabilities by emailing **security@corynth.io** (or create a GitHub Security Advisory for this repository).

**Please do not report security vulnerabilities through public GitHub issues.**

### What to Include

When reporting a vulnerability, please include:

- A description of the vulnerability
- Steps to reproduce the issue
- Potential impact of the vulnerability
- Any suggested fixes or mitigations

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Resolution**: Target within 30 days for critical vulnerabilities

### Security Features

Corynth includes several security features:

- **Input Validation**: All workflow parameters are validated before execution
- **Path Traversal Protection**: File operations are restricted to safe paths
- **Plugin Sandboxing**: Plugins run with limited system access
- **Secure Plugin Loading**: Plugins are compiled and loaded safely
- **Timeout Enforcement**: All operations have configurable timeouts

### Best Practices

When using Corynth:

- Use environment variables for sensitive data (API keys, passwords)
- Validate all external inputs in custom plugins
- Use specific versions for production workflows
- Regularly update Corynth and plugins
- Monitor workflow execution logs for suspicious activity
- Run workflows with minimal required privileges

## Vulnerability Disclosure Policy

We follow a responsible disclosure policy:

1. **Private disclosure** to our security team
2. **Confirmation** and impact assessment
3. **Fix development** and testing
4. **Public disclosure** after fix is available
5. **Security advisory** published with details

Thank you for helping keep Corynth secure!
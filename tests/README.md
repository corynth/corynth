# Corynth Tests

This directory contains tests for the Corynth project, including unit tests, integration tests, and end-to-end tests.

## Test Structure

- `unit/`: Unit tests for individual components
- `integration/`: Integration tests for component interactions
- `e2e/`: End-to-end tests for complete workflows
- `fixtures/`: Test fixtures and mock data

## Running Tests

To run all tests:

```bash
go test ./tests/...
```

To run a specific test category:

```bash
go test ./tests/unit/...
go test ./tests/integration/...
go test ./tests/e2e/...
```

To run a specific test file:

```bash
go test ./tests/unit/engine_test.go
```

## Writing Tests

When writing tests, follow these guidelines:

1. Use table-driven tests where appropriate
2. Mock external dependencies
3. Use descriptive test names
4. Include both positive and negative test cases
5. Keep tests independent and idempotent
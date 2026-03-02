# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a Contract Validator Toolkit (CVT) demonstration application showcasing runtime contract validation between API producers and consumers. It demonstrates three validation approaches: Manual, HTTP Adapter, and Mock Client.

**Architecture:**

- **CVT Server**: gRPC-based contract validation service (port 9550)
- **Producer**: Go-based Calculator API with CVT middleware (port 10001)
- **Consumers**: Node.js (consumer-1), Python (consumer-2), Java (consumer-3), and Go (consumer-4) CLI tools

## Common Commands

### Docker Operations

```bash
make up              # Start CVT server and producer
make down            # Stop all services
make logs            # View logs
make build           # Build all Docker images
```


### Testing

```bash
make test            # Run all tests (unit + integration)
make test-unit       # Unit/mock tests only (no producer required)
make test-integration # Integration tests (requires services running)

# Consumer-specific tests
make test-consumer-1      # All Consumer-1 tests
make test-consumer-2      # All Consumer-2 tests
make test-consumer-3      # All Consumer-3 tests (Java)
make test-consumer-4      # All Consumer-4 tests (Go)
make test-consumer-1-mock # Mock tests only

# Producer contract tests
make test-producer              # All producer tests
make test-producer-compliance   # Schema compliance tests
make test-producer-middleware   # Middleware mode tests
make test-producer-registry     # Consumer registry tests
make test-producer-integration  # HTTP integration tests
```

### Running Consumers

```bash
# Consumer-1 (Node.js) - add, subtract
make consumer-1-add x=5 y=3
make consumer-1-add-validate x=5 y=3  # With CVT validation

# Consumer-2 (Python) - add, multiply, divide
make consumer-2-multiply x=5 y=3
make consumer-2-divide-validate x=10 y=2

# Consumer-3 (Java) - multiply, divide
make consumer-3-multiply x=5 y=3
make consumer-3-multiply-validate x=5 y=3  # With CVT validation

# Consumer-4 (Go) - add, subtract
make consumer-4-add x=5 y=3
make consumer-4-subtract-validate x=10 y=2
```

### Direct Test Execution

```bash
# Consumer-1 (Node.js + Jest)
cd consumer-1 && npm test
npm run test:mock           # Mock tests only
npm run test:integration    # Integration tests

# Consumer-2 (Python + pytest)
cd consumer-2 && uv sync --extra dev --extra cvt
uv run pytest tests/ -v
uv run pytest tests/test_mock.py -v
```

### Demo Breaking Change Detection

```bash
make demo-breaking-change
```

## Architecture

### Three Validation Approaches

1. **Manual Validation** (`manual.test.js` / `test_manual.py`): Explicit `validator.validate()` calls with full control over validation

2. **HTTP Adapter** (`adapter.test.js` / `test_adapter.py`): Automatic validation via interceptors (axios interceptors for Node.js, `ContractValidatingSession` for Python)

3. **Mock Client** (`mock.test.js` / `test_mock.py`): Schema-generated responses without real HTTP calls - useful for unit testing in isolation (requires CVT server only)

### Producer Testing Approaches

1. **Schema Compliance** (`compliance_test.go`): ProducerTestKit validates handler responses against OpenAPI schema - unit testing without running server

2. **Middleware Modes** (`middleware_test.go`): Tests Strict/Warn/Shadow modes - Strict blocks invalid, Warn logs, Shadow collects metrics only

3. **Consumer Registry** (`registry_test.go`): Can-i-deploy checks verify schema changes won't break registered consumers

4. **HTTP Integration** (`integration_test.go`): Full HTTP tests against running producer with CVT validation

### Key Files

| Path                              | Purpose                                           |
| --------------------------------- | ------------------------------------------------- |
| `producer/main.go`                | Go Calculator API with CVT middleware             |
| `producer/handlers/calculator.go` | Extracted HTTP handlers with structured types     |
| `producer/calculator-api.yaml`    | OpenAPI 3.0.3 contract spec                       |
| `producer/tests/`                 | Producer contract test suites                     |
| `consumer-1/main.js`              | Node.js CLI with optional CVT validation          |
| `consumer-2/main.py`              | Python CLI with optional CVT validation           |
| `consumer-*/tests/`               | Consumer test suites for each validation approach |
| `consumer-3/src/main/java/demo/consumer3/Main.java` | Java CLI for multiply/divide |
| `consumer-3/src/test/java/demo/consumer3/` | Consumer-3 test suites          |
| `consumer-3/pom.xml`              | Maven build configuration                         |
| `consumer-4/main.go`              | Go CLI for add/subtract                           |
| `consumer-4/tests/`               | Consumer-4 test suites                            |
| `consumer-4/go.mod`               | Go module configuration                           |

### API Endpoints

All endpoints accept query params `x` and `y` (numbers) and return `{"result": <number>}`:

- `GET /add`, `/subtract`, `/multiply`, `/divide`
- `GET /health` returns `{"status": "healthy"}`

## SDK Dependencies

All CVT SDKs are consumed from published packages (v0.3.0):

- **Go SDK**: `github.com/sahina/cvt/sdks/go` via Go module proxy
- **Node SDK**: `@sahina/cvt-sdk` via npm public registry (`npmjs.org`)
- **Python SDK**: `cvt-sdk` via PyPI
- **Java SDK**: `io.github.sahina:cvt-sdk:0.3.0` (Maven Central)
- **CVT Server/CLI**: `ghcr.io/sahina/cvt:0.3.0` container image (unified binary)

## Environment Variables

| Variable          | Default                  | Purpose                               |
| ----------------- | ------------------------ | ------------------------------------- |
| `PRODUCER_URL`    | `http://localhost:10001` | Producer API URL                      |
| `CVT_SERVER_ADDR` | `localhost:9550`         | CVT gRPC server                       |
| `CVT_ENABLED`     | `true`                   | Enable/disable CVT on producer        |
| `CVT_ENVIRONMENT` | `demo`                   | Environment for consumer registration |

## Tool Versions

Managed via `.tool-versions` (asdf):

- Go 1.24.6
- Python 3.12.4
- Node 22.14.0
- Java 21 (via `eclipse-temurin:21` in Docker)

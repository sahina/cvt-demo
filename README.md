# CVT Demo Application

This demo application showcases the **[Contract Validator Toolkit (CVT)](https://github.com/sahina/cvt)** for API contract testing. It demonstrates how CVT enables runtime contract validation between API producers and consumers.

## Architecture

```text
                    ┌─────────────────┐
                    │   CVT Server    │
                    │   (port 9550)   │
                    └────────┬────────┘
                             │ gRPC
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│   Consumer-1    │ │    Producer     │ │   Consumer-2    │
│   (Node.js)     │ │   (Go + CVT)    │ │   (Python/uv)   │
│   CLI Tool      │ │  port 10001     │ │   CLI Tool      │
│  add, subtract  │ │  4 endpoints    │ │ add,mult,divide │
└─────────────────┘ └─────────────────┘ └─────────────────┘
         │                   ▲                   │
         └───────────────────┴───────────────────┘
                      HTTP calls
```

## Components

### Producer (Go)

A Go HTTP server implementing the Calculator API with 4 endpoints:

- `GET /add?x=<num>&y=<num>` - Add two numbers
- `GET /subtract?x=<num>&y=<num>` - Subtract two numbers
- `GET /multiply?x=<num>&y=<num>` - Multiply two numbers
- `GET /divide?x=<num>&y=<num>` - Divide two numbers

```bash
curl "http://localhost:10001/add?x=5&y=3"       # {"result":8}
curl "http://localhost:10001/subtract?x=10&y=4" # {"result":6}
curl "http://localhost:10001/multiply?x=4&y=7"  # {"result":28}
curl "http://localhost:10001/divide?x=20&y=4"   # {"result":5}
```

The producer uses CVT middleware for runtime contract validation against `calculator-api.yaml`.

### Consumer-1 (Node.js)

A CLI tool that calls the Calculator API for **add** and **subtract** operations:

```bash
node main.js add 5 3
node main.js subtract 10 4
```

### Consumer-2 (Python)

A CLI tool that calls the Calculator API for **add**, **multiply**, and **divide** operations:

```bash
uv run python main.py add 5 3
uv run python main.py multiply 4 7
uv run python main.py divide 10 2
```

## Prerequisites

- Docker and Docker Compose
- Make (optional, for convenience commands)
- curl and jq (for direct API testing)
- Access to the CVT repository (for SDKs and server)

## Quick Start

### 1. Clone the Repositories

This demo requires the CVT repository to be cloned alongside it (the docker-compose.yml mounts the CVT server and SDKs from the parent directory).

```bash
# Clone both repositories side by side
cd /path/to/your/workspace
git clone https://github.com/sahina/cvt.git
git clone https://github.com/sahina/cvt-demo.git

# Your directory structure should look like:
# workspace/
# ├── cvt/           # CVT server and SDKs
# └── cvt-demo/      # This demo application
```

### 2. Start the Infrastructure

```bash
cd cvt-demo

# Build and start CVT server + producer
make up

# Or without make:
docker compose up -d
```

### 3. Test the Producer Directly

```bash
# Quick health check
curl "http://localhost:10001/health"
# Expected: {"status":"healthy"}

# Test endpoints manually
curl "http://localhost:10001/add?x=5&y=3"
# Expected: {"result":8}

# Run producer contract tests (see Producer Contract Tests section)
make test-producer
```

### 4. Run Consumer-1 (Node.js)

```bash
# Without CVT validation (default: A=5, B=3)
make consumer-1-add
make consumer-1-subtract

# With custom values
make consumer-1-add x=10 y=20

# With CVT validation
make consumer-1-add-validate
make consumer-1-subtract-validate x=100 y=50
```

### 5. Run Consumer-2 (Python)

```bash
# Without CVT validation (default: A=5, B=3)
make consumer-2-add
make consumer-2-multiply
make consumer-2-divide

# With custom values
make consumer-2-multiply x=7 y=8
make consumer-2-divide x=100 y=4

# With CVT validation
make consumer-2-add-validate
make consumer-2-multiply-validate x=12 y=12
make consumer-2-divide-validate
```

### 6. Run All Tests

```bash
# Run all consumer operations
make test-all

# Run all with CVT validation
make test-contracts
```

### 7. Stop Everything

```bash
make down
```

## CVT Validation

Both consumers support optional CVT validation via the `--validate` flag:

### How It Works

**Runtime validation** (when running the CLI tools):

1. **Producer-side validation**: Every HTTP request/response is validated against the OpenAPI schema using CVT middleware as it passes through the producer server (enabled by default, can be disabled via `CVT_ENABLED=false`).

2. **Consumer-side validation** (optional): When `--validate` is enabled, consumers validate responses through the CVT server after each HTTP call.

**Test-time validation** (when running contract tests):

The contract tests in `consumer-1/tests/` and `consumer-2/tests/` use CVT to validate API interactions during test execution. See [Consumer Contract Tests](#consumer-contract-tests) for details on the three validation approaches (Manual, HTTP Adapter, and Mock Client).

### Enabling Validation

```bash
# Node.js consumer
node main.js add 5 3 --validate

# Python consumer
uv run python main.py add 5 3 --validate
```

When validation fails, the consumer exits with error code 1 and prints the validation errors.

## Consumer Contract Tests

This demo includes comprehensive contract tests demonstrating all three CVT validation approaches.

### Validation Approaches

| Approach         | Description                                                         | Services Required             |
| ---------------- | ------------------------------------------------------------------- | ----------------------------- |
| **Manual**       | Explicit `validator.validate()` calls with request/response objects | Producer running + CVT server |
| **HTTP Adapter** | Automatic validation via axios/requests interceptors                | Producer running + CVT server |
| **Mock Client**  | Schema-generated responses without real HTTP calls                  | CVT server only               |

- **Producer running**: The producer API must be accepting requests at `localhost:10001`
- **CVT server**: The CVT gRPC server must be running at `localhost:9550` (for schema registration and validation)
- Mock tests don't need the producer because CVT generates fake responses directly from the OpenAPI schema

### Running Contract Tests

```bash
# Install test dependencies
cd consumer-1 && npm install
cd consumer-2 && uv sync --extra dev --extra cvt

# Run mock tests (no producer needed, only CVT server)
make test-unit

# Run integration tests (requires producer + CVT server running)
make test-integration

# Run all tests for a specific consumer
make test-consumer-1
make test-consumer-2
```

### Test Files

For detailed documentation on each consumer's tests, see:

- [Consumer-1 Tests README](consumer-1/tests/README.md)
- [Consumer-2 Tests README](consumer-2/tests/README.md)

#### Consumer-1 Tests (Node.js)

- `manual.test.js` - Manual validation with explicit `validator.validate()` calls
- `adapter.test.js` - HTTP Adapter with automatic axios interceptors
- `mock.test.js` - Mock Client for unit testing (no producer needed)
- `registration.test.js` - Consumer registration (auto + manual)
- **Endpoints tested:** `/add`, `/subtract`

#### Consumer-2 Tests (Python)

- `test_manual.py` - Manual validation with explicit validate calls
- `test_adapter.py` - HTTP Adapter with `ContractValidatingSession`
- `test_mock.py` - Mock Client for unit testing (no producer needed)
- `test_registration.py` - Consumer registration (auto + manual)
- **Endpoints tested:** `/add`, `/multiply`, `/divide`

## Producer Contract Tests

This demo includes comprehensive producer-side contract tests demonstrating three CVT validation approaches.

### Producer Testing Approaches

| Approach              | Description                                                  | Services Required |
| --------------------- | ------------------------------------------------------------ | ----------------- |
| **Schema Compliance** | ProducerTestKit validates handler responses against schema   | CVT server only   |
| **Middleware Modes**  | Tests Strict/Warn/Shadow modes for runtime validation        | CVT server only   |
| **Consumer Registry** | Can-i-deploy checks verify changes won't break consumers     | CVT server only   |
| **HTTP Integration**  | Full HTTP tests against running producer with CVT validation | Producer + CVT    |

### Running Producer Tests

```bash
# Run all producer tests
make test-producer

# Run specific test types
make test-producer-compliance    # Schema compliance tests (unit)
make test-producer-middleware    # Middleware mode tests (unit)
make test-producer-registry      # Consumer registry / can-i-deploy tests
make test-producer-integration   # HTTP integration tests
```

### Producer Test Files

For detailed documentation, see [Producer Tests README](producer/tests/README.md).

- `compliance_test.go` - Schema compliance with ProducerTestKit
- `middleware_test.go` - Strict/Warn/Shadow mode testing
- `registry_test.go` - Can-i-deploy verification
- `integration_test.go` - Full HTTP integration tests

## Breaking Change Demo

This section demonstrates how CVT can detect breaking changes before they affect consumers.

### Scenario: Renaming a Response Field

1. **Initial state**: All consumers work with v1.0.0 of the API, which returns `{"result": <number>}`.

2. **Proposed change**: Rename `result` to `value` in v2.0.0 (see `producer/calculator-api-v2-breaking.yaml`).

3. **Impact analysis**: CVT detects that both consumers will break because they depend on the `result` field.

### Run the Demo

```bash
# Start infrastructure
make up

# Run the breaking change demo
make demo-breaking-change
```

### Manual Steps

```bash
# 1. Start the infrastructure
make up

# 2. Register consumers by running registration tests
make test-consumer-1-registration
make test-consumer-2-registration

# 3. Check which consumers would break with v2.0.0
# The v2 schema (calculator-api-v2-breaking.yaml) renames 'result' to 'value'
cvt can-i-deploy --schema calculator-api --version 2.0.0 --env demo

# Expected: UNSAFE - both consumers will break
#   - consumer-1 uses 'result' field in /add and /subtract
#   - consumer-2 uses 'result' field in /add, /multiply, and /divide
```

## API Reference

### Calculator API Endpoints

| Endpoint    | Method | Parameters         | Response                |
| ----------- | ------ | ------------------ | ----------------------- |
| `/add`      | GET    | `x`, `y` (numbers) | `{"result": <number>}`  |
| `/subtract` | GET    | `x`, `y` (numbers) | `{"result": <number>}`  |
| `/multiply` | GET    | `x`, `y` (numbers) | `{"result": <number>}`  |
| `/divide`   | GET    | `x`, `y` (numbers) | `{"result": <number>}`  |
| `/health`   | GET    | -                  | `{"status": "healthy"}` |

### Error Responses

All endpoints return 400 Bad Request for:

- Missing `x` or `y` parameters
- Non-numeric parameter values
- Division by zero (for `/divide`)

```json
{ "error": "error message" }
```

## Port Assignments

| Service              | Port  |
| -------------------- | ----- |
| CVT Server (gRPC)    | 9550  |
| CVT Server (Metrics) | 9551  |
| Producer             | 10001 |

## Make Targets

```bash
make help  # Show all available targets
```

### Docker Operations

- `make build` - Build all Docker images
- `make up` - Start CVT server and producer
- `make down` - Stop all services
- `make logs` - View logs from all services
- `make clean` - Remove containers and images

### Consumer Operations

- `make consumer-1-add` - Run add operation (Node.js)
- `make consumer-1-subtract` - Run subtract operation (Node.js)
- `make consumer-2-add` - Run add operation (Python)
- `make consumer-2-multiply` - Run multiply operation (Python)
- `make consumer-2-divide` - Run divide operation (Python)

Add `-validate` suffix for CVT validation (e.g., `make consumer-1-add-validate`).

**Custom values:** Pass `x` and `y` parameters to use custom numbers (default: x=5, y=3):

```bash
make consumer-1-add x=10 y=20        # 10 + 20 = 30
make consumer-2-multiply x=7 y=8     # 7 * 8 = 56
make consumer-2-divide x=100 y=4     # 100 / 4 = 25
```

### Testing

- `make test-all` - Run all consumer operations
- `make test-contracts` - Run all with CVT validation

### Consumer Contract Tests

- `make test-consumer-1` - Run all Consumer-1 tests
- `make test-consumer-1-mock` - Run Consumer-1 mock tests (no producer needed)
- `make test-consumer-1-integration` - Run Consumer-1 integration tests
- `make test-consumer-1-registration` - Run Consumer-1 registration tests
- `make test-consumer-2` - Run all Consumer-2 tests
- `make test-consumer-2-mock` - Run Consumer-2 mock tests (no producer needed)
- `make test-consumer-2-integration` - Run Consumer-2 integration tests
- `make test-consumer-2-registration` - Run Consumer-2 registration tests
- `make test-unit` - Run all mock/unit tests
- `make test-integration` - Run all integration tests
- `make demo-breaking-change` - Demo CVT breaking change detection

### Producer Contract Tests

- `make test-producer` - Run all producer tests
- `make test-producer-compliance` - Schema compliance tests (unit)
- `make test-producer-middleware` - Middleware mode tests (unit)
- `make test-producer-registry` - Consumer registry / can-i-deploy tests
- `make test-producer-integration` - HTTP integration tests

### Utilities

- `make shell-producer` - Shell into producer container
- `make shell-cvt` - Shell into CVT server container

## Project Structure

```text
cvt-demo/
├── docker-compose.yml     # Service orchestration
├── Makefile               # Convenience commands
├── README.md              # This file
├── producer/
│   ├── main.go            # Go HTTP server with CVT middleware
│   ├── go.mod             # Go module
│   ├── calculator-api.yaml # OpenAPI spec (v1.0.0)
│   ├── calculator-api-v2-breaking.yaml # Breaking schema (v2.0.0)
│   ├── Dockerfile
│   ├── handlers/
│   │   └── calculator.go  # HTTP handlers with structured types
│   └── tests/
│       ├── README.md      # Producer test documentation
│       ├── testutil_test.go # Shared test utilities
│       ├── compliance_test.go # Schema compliance tests
│       ├── middleware_test.go # Middleware mode tests
│       ├── registry_test.go # Consumer registry tests
│       └── integration_test.go # HTTP integration tests
├── consumer-1/
│   ├── main.js            # Node.js CLI
│   ├── package.json
│   ├── jest.config.js     # Jest configuration
│   ├── Dockerfile
│   └── tests/
│       ├── README.md      # Test documentation
│       ├── manual.test.js # Manual validation tests
│       ├── adapter.test.js # HTTP Adapter tests
│       ├── mock.test.js   # Mock Client tests
│       └── registration.test.js # Consumer registration tests
└── consumer-2/
    ├── main.py            # Python CLI
    ├── pyproject.toml
    ├── Dockerfile
    └── tests/
        ├── README.md      # Test documentation
        ├── conftest.py    # pytest fixtures
        ├── test_manual.py # Manual validation tests
        ├── test_adapter.py # HTTP Adapter tests
        ├── test_mock.py   # Mock Client tests
        └── test_registration.py # Consumer registration tests
```

## Development

### Local Development (without Docker)

**Producer:**

```bash
cd producer
go mod tidy
go run main.go

# Run producer tests (requires CVT server running)
go test ./tests/... -v
```

**Consumer-1:**

```bash
cd consumer-1
npm install
node main.js add 5 3

# Run tests
npm test                    # All tests
npm run test:mock           # Mock tests only
npm run test:integration    # Integration tests
```

**Consumer-2:**

```bash
cd consumer-2
uv sync --extra dev --extra cvt
uv run python main.py add 5 3

# Run tests
uv run pytest tests/ -v              # All tests
uv run pytest tests/test_mock.py -v  # Mock tests only
```

### Environment Variables

| Variable          | Default                  | Description                           |
| ----------------- | ------------------------ | ------------------------------------- |
| `PRODUCER_URL`    | `http://localhost:10001` | Producer API URL                      |
| `CVT_SERVER_ADDR` | `localhost:9550`         | CVT gRPC server address               |
| `SCHEMA_PATH`     | `./calculator-api.yaml`  | Path to OpenAPI schema                |
| `CVT_ENABLED`     | `true`                   | Enable/disable CVT on producer        |
| `CVT_ENVIRONMENT` | `demo`                   | Environment for consumer registration |

## Troubleshooting

### CVT Server Not Reachable

```bash
# Check if CVT server is running
docker compose logs cvt-server

# Check health
curl http://localhost:9551/health
```

### Producer Not Responding

```bash
# Check producer logs
docker compose logs producer

# Test health endpoint
curl http://localhost:10001/health
```

### Consumer Validation Failing

```bash
# Check if schema is registered correctly
docker compose logs cvt-server | grep "calculator-api"

# Try running without validation first
make consumer-1-add  # Should work
make consumer-1-add-validate  # Then try with validation
```

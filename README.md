# CVT Demo Application

This demo application showcases the **Contract Validator Toolkit (CVT)** for API contract testing. It demonstrates how CVT enables runtime contract validation between API producers and consumers.

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

- `GET /add?a=<num>&b=<num>` - Add two numbers
- `GET /subtract?a=<num>&b=<num>` - Subtract two numbers
- `GET /multiply?a=<num>&b=<num>` - Multiply two numbers
- `GET /divide?a=<num>&b=<num>` - Divide two numbers

```bash
curl "http://localhost:10001/add?a=5&b=3"       # {"result":8}
curl "http://localhost:10001/subtract?a=10&b=4" # {"result":6}
curl "http://localhost:10001/multiply?a=4&b=7"  # {"result":28}
curl "http://localhost:10001/divide?a=20&b=4"   # {"result":5}
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
# Test all endpoints
make test-producer

# Or manually:
curl "http://localhost:10001/add?a=5&b=3"
# Expected: {"result":8}
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

1. **Producer-side validation**: The producer always validates requests/responses against the OpenAPI schema using CVT middleware.

2. **Consumer-side validation** (optional): When `--validate` is enabled, consumers also validate responses through the CVT server.

### Enabling Validation

```bash
# Node.js consumer
node main.js add 5 3 --validate

# Python consumer
uv run python main.py add 5 3 --validate
```

When validation fails, the consumer exits with error code 1 and prints the validation errors.

## Breaking Change Demo

This section demonstrates how CVT can detect breaking changes before they affect consumers.

### Scenario: Removing an Endpoint

1. **Initial state**: All consumers work with v1.0.0 of the API.

2. **Proposed change**: Remove the `/subtract` endpoint in v2.0.0.

3. **Impact analysis**: CVT can detect that Consumer-1 (which uses `/subtract`) would break.

### Steps to Reproduce

```bash
# 1. Start the infrastructure
make up

# 2. Run consumer-1 (uses add and subtract)
make consumer-1-add-validate
make consumer-1-subtract-validate

# 3. Now imagine we want to remove /subtract endpoint
# CVT would detect this as a breaking change for Consumer-1

# 4. Run consumer-2 (uses add, multiply, divide - NOT subtract)
make consumer-2-add-validate
make consumer-2-multiply-validate
# Consumer-2 would be unaffected by removing /subtract
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
│   ├── main.go            # Go HTTP server
│   ├── go.mod             # Go module
│   ├── calculator-api.yaml # OpenAPI spec
│   └── Dockerfile
├── consumer-1/
│   ├── main.js            # Node.js CLI
│   ├── package.json
│   └── Dockerfile
└── consumer-2/
    ├── main.py            # Python CLI
    ├── pyproject.toml
    └── Dockerfile
```

## Development

### Local Development (without Docker)

**Producer:**

```bash
cd producer
go mod tidy
go run main.go
```

**Consumer-1:**

```bash
cd consumer-1
npm install
node main.js add 5 3
```

**Consumer-2:**

```bash
cd consumer-2
uv sync
uv run python main.py add 5 3
```

### Environment Variables

| Variable          | Default                  | Description                    |
| ----------------- | ------------------------ | ------------------------------ |
| `PRODUCER_URL`    | `http://localhost:10001` | Producer API URL               |
| `CVT_SERVER_ADDR` | `localhost:9550`         | CVT gRPC server address        |
| `SCHEMA_PATH`     | `./calculator-api.yaml`  | Path to OpenAPI schema         |
| `CVT_ENABLED`     | `true`                   | Enable/disable CVT on producer |

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

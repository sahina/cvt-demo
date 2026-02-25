# Design: Add Consumer-3 (Java) and Consumer-4 (Go)

**Date:** 2026-02-25
**Status:** Approved

## Overview

Add two new consumers to the CVT demo project:

- **consumer-3 (Java 21, Maven)** — supports `multiply` and `divide` operations
- **consumer-4 (Go 1.25)** — supports `add` and `subtract` operations

Both consumers follow Approach A (Full Parity): all 4 test types (manual, adapter, mock, registration), CLI entry point, Docker service, Makefile targets, and CI jobs — mirroring consumer-1 and consumer-2 exactly.

## Directory Structure

```
consumer-3/
├── pom.xml                                      # Maven build, Java 21, JUnit 5
├── Dockerfile                                   # Multi-stage: eclipse-temurin:21 → jre slim
├── src/
│   ├── main/java/demo/consumer3/
│   │   └── Main.java                            # CLI: multiply, divide [--validate]
│   └── test/java/demo/consumer3/
│       ├── ManualValidationTest.java            # Explicit validator.validate() calls
│       ├── AdapterValidationTest.java           # Auto-validation via SDK HTTP wrapper
│       ├── MockValidationTest.java              # Schema-generated responses, CVT only
│       └── RegistrationTest.java                # Register consumer, can-i-deploy checks

consumer-4/
├── go.mod                                       # Separate module, cvt/sdks/go@v0.3.0
├── Dockerfile                                   # Multi-stage: golang:1.25 → alpine:3.19
├── main.go                                      # CLI: add, subtract [--validate]
└── tests/
    ├── manual_test.go                           # Explicit Validate() calls
    ├── adapter_test.go                          # Auto-validation via SDK wrapper
    ├── mock_test.go                             # Schema-generated responses, CVT only
    └── registration_test.go                     # Register consumer, can-i-deploy checks
```

**Files modified (not created):**
- `docker-compose.yml` — add consumer-3 and consumer-4 services
- `Makefile` — add operation and test targets for both consumers
- `.github/workflows/test.yml` — add consumer-3-tests and consumer-4-tests jobs
- `.github/workflows/consumer-only-test.yml` — add mock-only jobs for both consumers
- `README.md` — update architecture diagram, components, quick start, CLI usage, test docs, Makefile targets, project structure, local development
- `CLAUDE.md` — update key files table and commands

## Consumer-3 (Java)

### CLI Interface
```
java -jar consumer3.jar multiply <x> <y> [--validate]
java -jar consumer3.jar divide <x> <y> [--validate]
```

### Dependencies (`pom.xml`)
```xml
<dependency>
    <groupId>io.github.sahina</groupId>
    <artifactId>cvt-sdk</artifactId>
    <version>0.3.0</version>
</dependency>
<!-- JUnit 5 for testing -->
<dependency>
    <groupId>org.junit.jupiter</groupId>
    <artifactId>junit-jupiter</artifactId>
    <version>5.x</version>
    <scope>test</scope>
</dependency>
```

HTTP calls use Java 21's built-in `HttpClient` — no extra dependencies.

### Test Structure

| Class | Approach | Requires |
|---|---|---|
| `ManualValidationTest` | Explicit `validator.validate(request, response)` | CVT + Producer |
| `AdapterValidationTest` | SDK HTTP wrapper with auto-validation | CVT + Producer |
| `MockValidationTest` | Schema-generated responses | CVT only |
| `RegistrationTest` | `registerConsumerFromInteractions()` + can-i-deploy | CVT only |

### Maven Test Execution
```bash
mvn test -Dtest="MockValidationTest"                      # mock only
mvn test -Dtest="ManualValidationTest,AdapterValidationTest"  # live tests
mvn test                                                   # all tests
```

### Dockerfile
Multi-stage build:
1. `eclipse-temurin:21` — compile and package with `mvn package -DskipTests`
2. `eclipse-temurin:21-jre` — runtime with JAR + `calculator-api.json`

### SDK API Note
The exact class/method names for `io.github.sahina:cvt-sdk:0.3.0` will be verified against the published JAR during implementation, using the same conceptual patterns as the Node.js and Python SDKs (`ContractValidator`, adapters, mock client).

## Consumer-4 (Go)

### CLI Interface
```
./consumer4 add <x> <y> [--validate]
./consumer4 subtract <x> <y> [--validate]
```

### Module (`go.mod`)
```
module github.com/sahina/cvt-demo/consumer-4

require github.com/sahina/cvt/sdks/go v0.3.0
```

Uses Go's standard `flag` package for CLI parsing and `net/http` for HTTP calls — no extra dependencies beyond the CVT SDK.

### Test Structure

| File | Approach | Requires |
|---|---|---|
| `manual_test.go` | Explicit `validator.Validate(req, resp)` | CVT + Producer |
| `adapter_test.go` | SDK HTTP wrapper with auto-validation | CVT + Producer |
| `mock_test.go` | Schema-generated responses | CVT only |
| `registration_test.go` | `RegisterConsumerFromInteractions()` + can-i-deploy | CVT only |

### Go Test Execution
```bash
go test ./tests/... -run TestMock -v                       # mock only
go test ./tests/... -run "TestManual|TestAdapter" -v       # live tests
go test ./tests/... -v                                     # all tests
```

### Dockerfile
Multi-stage build (same pattern as producer):
1. `golang:1.25-alpine` — `go build -o consumer4 .`
2. `alpine:3.19` — binary + `calculator-api.json`

## Makefile Additions

```makefile
# Consumer-3 (Java) — operations
consumer-3-multiply:          docker compose run --rm consumer-3 multiply $(x) $(y)
consumer-3-divide:            docker compose run --rm consumer-3 divide $(x) $(y)
consumer-3-multiply-validate: docker compose run --rm consumer-3 multiply $(x) $(y) --validate
consumer-3-divide-validate:   docker compose run --rm consumer-3 divide $(x) $(y) --validate

# Consumer-4 (Go) — operations
consumer-4-add:               docker compose run --rm consumer-4 add $(x) $(y)
consumer-4-subtract:          docker compose run --rm consumer-4 subtract $(x) $(y)
consumer-4-add-validate:      docker compose run --rm consumer-4 add $(x) $(y) --validate
consumer-4-subtract-validate: docker compose run --rm consumer-4 subtract $(x) $(y) --validate

# Consumer-3 tests
test-consumer-3:              cd consumer-3 && mvn test
test-consumer-3-mock:         cd consumer-3 && mvn test -Dtest="MockValidationTest"
test-consumer-3-integration:  cd consumer-3 && mvn test -Dtest="ManualValidationTest,AdapterValidationTest"
test-consumer-3-registration: cd consumer-3 && mvn test -Dtest="RegistrationTest"

# Consumer-4 tests
test-consumer-4:              cd consumer-4 && go test ./tests/... -v
test-consumer-4-mock:         cd consumer-4 && go test ./tests/... -run TestMock -v
test-consumer-4-integration:  cd consumer-4 && go test ./tests/... -run "TestManual|TestAdapter" -v
test-consumer-4-registration: cd consumer-4 && go test ./tests/... -run TestRegistration -v
```

`test-unit` and `test-integration` aggregate targets expand to include consumer-3 and consumer-4.

## Docker Compose Additions

```yaml
consumer-3:
  build:
    context: consumer-3
  depends_on: [producer]
  profiles: [cli]
  networks: [cvt-demo-network]
  environment:
    PRODUCER_URL: http://producer:10001
    CVT_SERVER_ADDR: cvt-server:9550
    SCHEMA_PATH: /app/calculator-api.json

consumer-4:
  build:
    context: consumer-4
  depends_on: [producer]
  profiles: [cli]
  networks: [cvt-demo-network]
  environment:
    PRODUCER_URL: http://producer:10001
    CVT_SERVER_ADDR: cvt-server:9550
    SCHEMA_PATH: /app/calculator-api.json
```

## GitHub Actions Additions

### `test.yml` — parallel fan-out after producer-tests

```
producer-tests → (consumer-1-tests || consumer-2-tests || consumer-3-tests || consumer-4-tests)
```

**consumer-3-tests job:**
- Uses `actions/setup-java@v4` (Java 21, temurin distribution)
- Phase 1 (CVT only): `mvn test -Dtest="MockValidationTest"`
- Phase 2 (CVT + Producer): `mvn test -Dtest="ManualValidationTest,AdapterValidationTest,RegistrationTest"`
- Same job summary format as existing consumer jobs

**consumer-4-tests job:**
- Uses `actions/setup-go@v5` (Go 1.25.x)
- Phase 1 (CVT only): `go test ./tests/... -run TestMock -v`
- Phase 2 (CVT + Producer): `go test ./tests/... -run "TestManual|TestAdapter|TestRegistration" -v`
- Same job summary format as existing consumer jobs

### `consumer-only-test.yml` — mock-only parallel jobs

Adds consumer-3 mock (`mvn test -Dtest=MockValidationTest`) and consumer-4 mock (`go test -run TestMock`) as additional parallel jobs alongside existing consumer-1 and consumer-2 mock jobs.

## Operation Assignment Summary

| Consumer | Language | Operations | Build Tool | Test Framework |
|---|---|---|---|---|
| consumer-1 | Node.js 22 | add, subtract | npm/Jest | Jest |
| consumer-2 | Python 3.12 | add, multiply, divide | uv/pytest | pytest |
| consumer-3 | Java 21 | multiply, divide | Maven | JUnit 5 |
| consumer-4 | Go 1.25 | add, subtract | go test | go test |

.PHONY: help build up down logs clean test-all test test-contracts \
	producer-up producer-down \
	consumer-1-add consumer-1-subtract \
	consumer-1-add-validate consumer-1-subtract-validate \
	consumer-2-add consumer-2-multiply consumer-2-divide \
	consumer-2-add-validate consumer-2-multiply-validate consumer-2-divide-validate \
	shell-producer shell-cvt test-producer-http \
	test-consumer-1 test-consumer-1-mock test-consumer-1-live test-consumer-1-registration \
	test-consumer-2 test-consumer-2-mock test-consumer-2-live test-consumer-2-registration \
	test-unit test-live demo-breaking-change \
	test-producer test-producer-compliance test-producer-middleware \
	test-producer-registry test-producer-integration

# =============================================================================
# Path Configuration for Docker Builds
# =============================================================================
# CVT repository root (relative to this repo)
CVT_ROOT ?= ../cvt

# Docker build context directory (parent of this repo)
BUILD_CONTEXT ?= ..

# Symlink names expected by Dockerfiles
CVT_SYMLINK_NAME ?= cvt
REPO_SYMLINK_NAME ?= cvt-demo

# Default values for calculator operations
x ?= 5
y ?= 3

# Default target
help:
	@echo "CVT Demo Application"
	@echo ""
	@echo "Docker Operations:"
	@echo "  make build              - Build all Docker images"
	@echo "  make up                 - Start CVT server and producer"
	@echo "  make down               - Stop all services"
	@echo "  make producer-up        - Start just the producer (CVT server must be running)"
	@echo "  make producer-down      - Stop just the producer"
	@echo "  make logs               - View logs from all services"
	@echo "  make clean              - Remove containers and images"
	@echo ""
	@echo "Producer Contract Tests:"
	@echo "  make test-producer           - Run all producer contract tests"
	@echo "  make test-producer-compliance - Schema compliance tests (no producer needed)"
	@echo "  make test-producer-middleware - Middleware mode tests (no producer needed)"
	@echo "  make test-producer-registry  - Consumer registry tests"
	@echo "  make test-producer-integration - HTTP integration tests (requires producer)"
	@echo "  make test-producer-http      - Quick HTTP endpoint test with curl"
	@echo ""
	@echo "Consumer-1 Operations (Node.js - add, subtract):"
	@echo "  make consumer-1-add             - Run: add (default: 5 + 3)"
	@echo "  make consumer-1-subtract        - Run: subtract (default: 5 - 3)"
	@echo "  make consumer-1-add-validate    - With CVT validation"
	@echo "  make consumer-1-subtract-validate - With CVT validation"
	@echo ""
	@echo "Consumer-2 Operations (Python - add, multiply, divide):"
	@echo "  make consumer-2-add             - Run: add (default: 5 + 3)"
	@echo "  make consumer-2-multiply        - Run: multiply (default: 5 * 3)"
	@echo "  make consumer-2-divide          - Run: divide (default: 5 / 3)"
	@echo "  make consumer-2-add-validate    - With CVT validation"
	@echo "  make consumer-2-multiply-validate - With CVT validation"
	@echo "  make consumer-2-divide-validate - With CVT validation"
	@echo ""
	@echo "Custom values: make <target> x=<num> y=<num>"
	@echo "  Example: make consumer-1-add x=10 y=20"
	@echo ""
	@echo "Testing:"
	@echo "  make test-all           - Run all consumer operations"
	@echo "  make test-contracts     - Run all consumers with CVT validation"
	@echo ""
	@echo "Consumer Contract Tests:"
	@echo "  make test-consumer-1           - Run all Consumer-1 tests"
	@echo "  make test-consumer-1-mock      - Run Consumer-1 mock tests (no producer needed)"
	@echo "  make test-consumer-1-live      - Run Consumer-1 live tests (requires producer)"
	@echo "  make test-consumer-1-registration - Run Consumer-1 registration tests"
	@echo "  make test-consumer-2           - Run all Consumer-2 tests"
	@echo "  make test-consumer-2-mock      - Run Consumer-2 mock tests (no producer needed)"
	@echo "  make test-consumer-2-live      - Run Consumer-2 live tests (requires producer)"
	@echo "  make test-consumer-2-registration - Run Consumer-2 registration tests"
	@echo "  make test-unit          - Run all mock tests (no producer needed)"
	@echo "  make test-live          - Run all live tests (requires producer)"
	@echo "  make test               - Run all tests (mock + live)"
	@echo ""
	@echo "Breaking Change Demo:"
	@echo "  make demo-breaking-change - Demo CVT breaking change detection"
	@echo ""
	@echo "Utilities:"
	@echo "  make shell-producer     - Shell into producer container"
	@echo "  make shell-cvt          - Shell into CVT server container"

# =============================================================================
# Docker Operations
# =============================================================================

# Check that CVT SDKs are available for Docker builds
check-cvt-sdk:
	@if [ ! -d "$(CVT_ROOT)/sdks/go" ] || [ ! -d "$(CVT_ROOT)/sdks/node" ] || [ ! -d "$(CVT_ROOT)/sdks/python" ]; then \
		echo ""; \
		echo "ERROR: CVT SDKs not found at $(CVT_ROOT)/sdks/{go,node,python}"; \
		echo ""; \
		echo "To build Docker images locally, clone the CVT repository:"; \
		echo "  git clone https://github.com/sahina/cvt.git $(CVT_ROOT)"; \
		echo ""; \
		exit 1; \
	fi

# Create symlinks in build context for Docker builds
# Dockerfiles expect: $(CVT_SYMLINK_NAME)/sdks/... and $(REPO_SYMLINK_NAME)/...
setup-symlink: check-cvt-sdk
	@# Create CVT symlink in build context pointing to CVT repository
	@if [ ! -e $(BUILD_CONTEXT)/$(CVT_SYMLINK_NAME) ]; then \
		ln -sf $(CVT_ROOT) $(BUILD_CONTEXT)/$(CVT_SYMLINK_NAME) 2>/dev/null || true; \
	fi
	@# Create repo symlink in build context pointing to current directory
	@DIRNAME=$$(basename $$(pwd)); \
	if [ "$$DIRNAME" != "$(REPO_SYMLINK_NAME)" ] && [ ! -e $(BUILD_CONTEXT)/$(REPO_SYMLINK_NAME) ]; then \
		ln -sf "$$DIRNAME" $(BUILD_CONTEXT)/$(REPO_SYMLINK_NAME) 2>/dev/null || true; \
	fi

# Local Docker build (requires CVT repository and symlinks)
build: setup-symlink
	docker compose build

up: build
	docker compose up -d

down:
	docker compose down

producer-up:
	@echo "Starting producer service..."
	docker compose up -d producer
	@echo "Producer is running at http://localhost:10001"

producer-down:
	@echo "Stopping producer service..."
	docker compose stop producer
	@echo "Producer stopped"

logs:
	docker compose logs -f

clean:
	docker compose down -v --rmi local
	docker compose --profile cli down -v --rmi local

# =============================================================================
# Producer Testing (direct HTTP)
# =============================================================================

test-producer-http:
	@echo "Testing producer endpoints with curl..."
	@echo ""
	@echo "GET /add?x=5&y=3"
	@curl -s "http://localhost:10001/add?x=5&y=3" | jq .
	@echo ""
	@echo "GET /subtract?x=10&y=4"
	@curl -s "http://localhost:10001/subtract?x=10&y=4" | jq .
	@echo ""
	@echo "GET /multiply?x=4&y=7"
	@curl -s "http://localhost:10001/multiply?x=4&y=7" | jq .
	@echo ""
	@echo "GET /divide?x=10&y=2"
	@curl -s "http://localhost:10001/divide?x=10&y=2" | jq .
	@echo ""
	@echo "GET /health"
	@curl -s "http://localhost:10001/health" | jq .

# =============================================================================
# Producer Contract Tests
# =============================================================================

test-producer-compliance:
	@echo "Running Producer schema compliance tests..."
	cd producer && go test ./tests/... -run Compliance -v

test-producer-middleware:
	@echo "Running Producer middleware mode tests..."
	cd producer && go test ./tests/... -run Middleware -v

test-producer-registry:
	@echo "Running Producer consumer registry tests..."
	cd producer && go test ./tests/... -run Registry -v

test-producer-integration:
	@echo "Running Producer HTTP integration tests (requires producer running)..."
	cd producer && go test ./tests/... -run Integration -v

test-producer:
	@echo "============================================"
	@echo "Running All Producer Contract Tests"
	@echo "============================================"
	@echo ""
	cd producer && go test ./tests/... -v
	@echo ""
	@echo "============================================"
	@echo "All producer tests completed!"
	@echo "============================================"

# =============================================================================
# Consumer-1 Operations (without validation)
# Usage: make consumer-1-add x=4 y=5
# =============================================================================

consumer-1-add:
	docker compose run --rm consumer-1 add $(x) $(y)

consumer-1-subtract:
	docker compose run --rm consumer-1 subtract $(x) $(y)

# =============================================================================
# Consumer-1 Operations (with CVT validation)
# Usage: make consumer-1-add-validate x=4 y=5
# =============================================================================

consumer-1-add-validate:
	docker compose run --rm consumer-1 add $(x) $(y) --validate

consumer-1-subtract-validate:
	docker compose run --rm consumer-1 subtract $(x) $(y) --validate

# =============================================================================
# Consumer-2 Operations (without validation)
# Usage: make consumer-2-multiply x=4 y=7
# =============================================================================

consumer-2-add:
	docker compose run --rm consumer-2 add $(x) $(y)

consumer-2-multiply:
	docker compose run --rm consumer-2 multiply $(x) $(y)

consumer-2-divide:
	docker compose run --rm consumer-2 divide $(x) $(y)

# =============================================================================
# Consumer-2 Operations (with CVT validation)
# Usage: make consumer-2-multiply-validate x=4 y=7
# =============================================================================

consumer-2-add-validate:
	docker compose run --rm consumer-2 add $(x) $(y) --validate

consumer-2-multiply-validate:
	docker compose run --rm consumer-2 multiply $(x) $(y) --validate

consumer-2-divide-validate:
	docker compose run --rm consumer-2 divide $(x) $(y) --validate

# =============================================================================
# Testing
# =============================================================================

test-all:
	@echo "Running all consumer operations..."
	@echo ""
	@echo "=== Consumer-1 (Node.js) ==="
	@$(MAKE) -s consumer-1-add
	@$(MAKE) -s consumer-1-subtract
	@echo ""
	@echo "=== Consumer-2 (Python) ==="
	@$(MAKE) -s consumer-2-add
	@$(MAKE) -s consumer-2-multiply
	@$(MAKE) -s consumer-2-divide
	@echo ""
	@echo "All operations completed!"

test-contracts:
	@echo "Running all consumer operations with CVT validation..."
	@echo ""
	@echo "=== Consumer-1 (Node.js) with validation ==="
	@$(MAKE) -s consumer-1-add-validate
	@$(MAKE) -s consumer-1-subtract-validate
	@echo ""
	@echo "=== Consumer-2 (Python) with validation ==="
	@$(MAKE) -s consumer-2-add-validate
	@$(MAKE) -s consumer-2-multiply-validate
	@$(MAKE) -s consumer-2-divide-validate
	@echo ""
	@echo "All contract validations passed!"

# =============================================================================
# Utilities
# =============================================================================

shell-producer:
	docker compose exec producer sh

shell-cvt:
	docker compose exec cvt-server sh

# =============================================================================
# Consumer Contract Tests
# =============================================================================

# Consumer-1 (Node.js) Tests
test-consumer-1-mock:
	@echo "Running Consumer-1 mock tests (no producer needed)..."
	cd consumer-1 && npm test -- --testPathPattern=mock

test-consumer-1-live:
	@echo "Running Consumer-1 live tests (requires producer)..."
	cd consumer-1 && npm test -- --testPathPattern="(adapter|manual)"

test-consumer-1-registration:
	@echo "Running Consumer-1 registration tests..."
	cd consumer-1 && npm test -- --testPathPattern=registration

test-consumer-1:
	@echo "Running all Consumer-1 tests..."
	cd consumer-1 && npm test

# Consumer-2 (Python) Tests
test-consumer-2-mock:
	@echo "Running Consumer-2 mock tests (no producer needed)..."
	cd consumer-2 && uv run pytest tests/test_mock.py -v

test-consumer-2-live:
	@echo "Running Consumer-2 live tests (requires producer)..."
	cd consumer-2 && uv run pytest tests/test_adapter.py tests/test_manual.py -v

test-consumer-2-registration:
	@echo "Running Consumer-2 registration tests..."
	cd consumer-2 && uv run pytest tests/test_registration.py -v

test-consumer-2:
	@echo "Running all Consumer-2 tests..."
	cd consumer-2 && uv run pytest tests/ -v

# Combined Test Targets
test-unit:
	@echo "Running all mock/unit tests (no services needed except CVT server)..."
	@$(MAKE) -s test-consumer-1-mock
	@$(MAKE) -s test-consumer-2-mock
	@echo ""
	@echo "All mock tests completed!"

test-live:
	@echo "Running all live tests (requires producer)..."
	@$(MAKE) -s test-consumer-1-live
	@$(MAKE) -s test-consumer-2-live
	@echo ""
	@echo "All live tests completed!"

# Run all tests (mock + live)
test:
	@echo "============================================"
	@echo "Running All Consumer Contract Tests"
	@echo "============================================"
	@echo ""
	@$(MAKE) -s test-unit
	@echo ""
	@$(MAKE) -s test-live
	@echo ""
	@echo "============================================"
	@echo "All tests completed!"
	@echo "============================================"

# =============================================================================
# Breaking Change Demo
# =============================================================================

demo-breaking-change:
	@echo "============================================"
	@echo "Breaking Change Detection Demo"
	@echo "============================================"
	@echo ""
	@echo "Step 1: Register consumers to 'demo' environment..."
	@echo "Running consumer-1 registration tests..."
	-cd consumer-1 && CVT_ENVIRONMENT=demo npm test -- --testPathPattern=registration 2>/dev/null || true
	@echo ""
	@echo "Running consumer-2 registration tests..."
	-cd consumer-2 && CVT_ENVIRONMENT=demo uv run pytest tests/test_registration.py -v 2>/dev/null || true
	@echo ""
	@echo "Step 2: Register v2.0.0 breaking schema..."
	@echo "(This schema changes 'result' field to 'value')"
	@echo ""
	@echo "Step 3: Check if v2.0.0 can be deployed..."
	@echo "cvt can-i-deploy --schema calculator-api --version 2.0.0 --env demo"
	@echo ""
	@echo "Expected result: UNSAFE - both consumers will break"
	@echo "  - consumer-1 uses 'result' field in /add and /subtract"
	@echo "  - consumer-2 uses 'result' field in /add, /multiply, and /divide"
	@echo ""
	@echo "The breaking change: 'result' field renamed to 'value' in v2.0.0"
	@echo "============================================"

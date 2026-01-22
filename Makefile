.PHONY: help build up down logs clean test-all \
	consumer-1-add consumer-1-subtract \
	consumer-1-add-validate consumer-1-subtract-validate \
	consumer-2-add consumer-2-multiply consumer-2-divide \
	consumer-2-add-validate consumer-2-multiply-validate consumer-2-divide-validate \
	shell-producer shell-cvt test-producer

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
	@echo "  make logs               - View logs from all services"
	@echo "  make clean              - Remove containers and images"
	@echo ""
	@echo "Producer Testing (direct HTTP):"
	@echo "  make test-producer      - Test producer endpoints directly"
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
	@echo "Utilities:"
	@echo "  make shell-producer     - Shell into producer container"
	@echo "  make shell-cvt          - Shell into CVT server container"

# =============================================================================
# Docker Operations
# =============================================================================

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	docker compose down -v --rmi local
	docker compose --profile cli down -v --rmi local

# =============================================================================
# Producer Testing (direct HTTP)
# =============================================================================

test-producer:
	@echo "Testing producer endpoints..."
	@echo ""
	@echo "GET /add?a=5&b=3"
	@curl -s "http://localhost:10001/add?a=5&b=3" | jq .
	@echo ""
	@echo "GET /subtract?a=10&b=4"
	@curl -s "http://localhost:10001/subtract?a=10&b=4" | jq .
	@echo ""
	@echo "GET /multiply?a=4&b=7"
	@curl -s "http://localhost:10001/multiply?a=4&b=7" | jq .
	@echo ""
	@echo "GET /divide?a=10&b=2"
	@curl -s "http://localhost:10001/divide?a=10&b=2" | jq .
	@echo ""
	@echo "GET /health"
	@curl -s "http://localhost:10001/health" | jq .

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

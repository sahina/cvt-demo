# Consumer-2 Dockerfile
FROM python:3.12-slim

WORKDIR /app

# Install uv for dependency management
RUN pip install --no-cache-dir uv==0.9.26

# Copy project config and install dependencies
COPY consumer-2/pyproject.toml ./
RUN uv sync --extra cvt

# Copy consumer source
COPY consumer-2/main.py ./

# Copy the OpenAPI schema for validation (JSON format required by CVT server)
COPY producer/calculator-api.json ./calculator-api.json

# Set environment variables
ENV PRODUCER_URL=http://producer:10001
ENV CVT_SERVER_ADDR=cvt-server:9550
ENV SCHEMA_PATH=/app/calculator-api.json

ENTRYPOINT ["uv", "run", "python", "main.py"]

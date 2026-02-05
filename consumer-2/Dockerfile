# Consumer-2 Dockerfile
FROM python:3.12-slim

WORKDIR /app

# Install uv for dependency management
RUN pip install --no-cache-dir uv==0.9.26

# Copy CVT SDK
COPY cvt/sdks/python /cvt/sdks/python

# Copy CVT proto files (required for gRPC communication)
COPY cvt/api/protos /cvt/api/protos

# Copy consumer source
COPY cvt-demo/consumer-2/pyproject.toml ./
COPY cvt-demo/consumer-2/main.py ./

# Copy the OpenAPI schema for validation (JSON format required by CVT server)
COPY cvt-demo/producer/calculator-api.json ./calculator-api.json

# Update the SDK path for Docker context and install dependencies
RUN sed -i 's|../../cvt/sdks/python|/cvt/sdks/python|g' pyproject.toml && \
  uv sync --extra cvt

# Set environment variables
ENV PRODUCER_URL=http://producer:10001
ENV CVT_SERVER_ADDR=cvt-server:9550
ENV SCHEMA_PATH=/app/calculator-api.json

ENTRYPOINT ["uv", "run", "python", "main.py"]

# Consumer-2 Dockerfile
FROM python:3.12-slim

WORKDIR /app

# Install uv for dependency management
RUN pip install --no-cache-dir uv

# Copy CVT SDK
COPY cvt/sdks/python /cvt/sdks/python

# Copy consumer source
COPY cvt-demo/consumer-2/pyproject.toml ./
COPY cvt-demo/consumer-2/main.py ./

# Copy the OpenAPI schema for validation
COPY cvt-demo/producer/calculator-api.yaml ./calculator-api.yaml

# Update the SDK path for Docker context
RUN sed -i 's|../../cvt/sdks/python|/cvt/sdks/python|g' pyproject.toml

# Install dependencies with uv
RUN uv pip install --system -e . && \
    uv pip install --system -e /cvt/sdks/python

# Set environment variables
ENV PRODUCER_URL=http://producer:10001
ENV CVT_SERVER_ADDR=cvt-server:9550
ENV SCHEMA_PATH=/app/calculator-api.yaml

ENTRYPOINT ["python", "main.py"]

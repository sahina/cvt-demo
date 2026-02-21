FROM node:20-alpine

WORKDIR /app

# Copy package files
COPY consumer-1/package.json consumer-1/package-lock.json ./

# Install dependencies
RUN npm ci

# Copy consumer source
COPY consumer-1/main.js ./

# Copy the OpenAPI schema for validation (JSON format required by CVT server)
COPY producer/calculator-api.json ./calculator-api.json

# Set environment variables
ENV PRODUCER_URL=http://producer:10001
ENV CVT_SERVER_ADDR=cvt-server:9550
ENV SCHEMA_PATH=/app/calculator-api.json

ENTRYPOINT ["node", "main.js"]

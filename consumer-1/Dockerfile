# Consumer-1 Dockerfile
FROM node:20-alpine

WORKDIR /app

# Copy CVT SDK source (dist is excluded by .dockerignore, so we build it)
COPY cvt/sdks/node /cvt/sdks/node

# Copy CVT proto files (required for gRPC communication)
COPY cvt/api/protos /cvt/api/protos

# Build the CVT SDK (TypeScript -> JavaScript)
RUN cd /cvt/sdks/node && npm install && npm run build

# Copy consumer source
COPY cvt-demo/consumer-1/package.json ./
COPY cvt-demo/consumer-1/main.js ./

# Copy the OpenAPI schema for validation (JSON format required by CVT server)
COPY cvt-demo/producer/calculator-api.json ./calculator-api.json

# Update the SDK path for Docker context and install dependencies
RUN sed -i 's|file:../../cvt/sdks/node|file:/cvt/sdks/node|g' package.json && \
    npm install

# Set environment variables
ENV PRODUCER_URL=http://producer:10001
ENV CVT_SERVER_ADDR=cvt-server:9550
ENV SCHEMA_PATH=/app/calculator-api.json

ENTRYPOINT ["node", "main.js"]

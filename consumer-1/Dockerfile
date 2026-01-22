# Consumer-1 Dockerfile
FROM node:20-alpine

WORKDIR /app

# Copy CVT SDK
COPY cvt/sdks/node /cvt/sdks/node

# Copy consumer source
COPY cvt-demo/consumer-1/package.json ./
COPY cvt-demo/consumer-1/main.js ./

# Copy the OpenAPI schema for validation
COPY cvt-demo/producer/calculator-api.yaml ./calculator-api.yaml

# Update the SDK path for Docker context
RUN sed -i 's|file:../../cvt/sdks/node|file:/cvt/sdks/node|g' package.json

# Install dependencies
RUN npm install

# Set environment variables
ENV PRODUCER_URL=http://producer:10001
ENV CVT_SERVER_ADDR=cvt-server:9550
ENV SCHEMA_PATH=/app/calculator-api.yaml

ENTRYPOINT ["node", "main.js"]

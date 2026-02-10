FROM node:20-alpine

WORKDIR /app

# Copy package files and .npmrc for GitHub packages registry
COPY consumer-1/.npmrc consumer-1/package.json consumer-1/package-lock.json ./

# Install dependencies (auth token injected via BuildKit secret)
RUN --mount=type=secret,id=github_token \
    echo "//npm.pkg.github.com/:_authToken=$(cat /run/secrets/github_token)" >> .npmrc && \
    npm ci && \
    sed -i '/authToken/d' .npmrc

# Copy consumer source
COPY consumer-1/main.js ./

# Copy the OpenAPI schema for validation (JSON format required by CVT server)
COPY producer/calculator-api.json ./calculator-api.json

# Set environment variables
ENV PRODUCER_URL=http://producer:10001
ENV CVT_SERVER_ADDR=cvt-server:9550
ENV SCHEMA_PATH=/app/calculator-api.json

ENTRYPOINT ["node", "main.js"]

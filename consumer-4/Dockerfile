# Consumer-4 Dockerfile
# Build context: repo root (docker build -f consumer-4/Dockerfile .)

# Build stage
FROM golang:1.25-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

WORKDIR /build

# Copy go module files and download dependencies (layer cache)
COPY consumer-4/go.mod consumer-4/go.sum ./
RUN go mod download

# Copy source and build
COPY consumer-4/*.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o consumer4 .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Copy binary
COPY --from=builder /build/consumer4 .

# Copy the OpenAPI schema for validation (JSON format required by CVT server)
COPY producer/calculator-api.json ./calculator-api.json

# Set environment variables
ENV PRODUCER_URL=http://producer:10001
ENV CVT_SERVER_ADDR=cvt:9550
ENV SCHEMA_PATH=/app/calculator-api.json

ENTRYPOINT ["./consumer4"]

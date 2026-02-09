# Build stage
FROM golang:1.25-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

WORKDIR /build

# Copy go module files and download dependencies
COPY producer/go.mod producer/go.sum ./
RUN go mod download

# Copy source code
COPY producer/ .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /calculator-api .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Copy binary and schema
COPY --from=builder /calculator-api .
COPY --from=builder /build/calculator-api.yaml .

# Set default environment variables
ENV PORT=10001
ENV CVT_SERVER_ADDR=cvt-server:9550
ENV SCHEMA_PATH=/app/calculator-api.yaml
ENV CVT_ENABLED=true

EXPOSE 10001

CMD ["./calculator-api"]

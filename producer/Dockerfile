# Build stage
FROM golang:1.25-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

WORKDIR /build

# Copy CVT SDK first
COPY cvt/sdks/go /cvt/sdks/go

# Copy the cvt-demo directory
COPY cvt-demo /build/cvt-demo

WORKDIR /build/cvt-demo/producer

# Update replace directive for Docker context
RUN sed -i 's|../../cvt/sdks/go|/cvt/sdks/go|g' go.mod

# Download dependencies and build
RUN go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /calculator-api .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Copy binary and schema
COPY --from=builder /calculator-api .
COPY --from=builder /build/cvt-demo/producer/calculator-api.yaml .

# Set default environment variables
ENV PORT=10001
ENV CVT_SERVER_ADDR=cvt-server:9550
ENV SCHEMA_PATH=/app/calculator-api.yaml
ENV CVT_ENABLED=true

EXPOSE 10001

CMD ["./calculator-api"]

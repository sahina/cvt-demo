# Consumer-3 Dockerfile
# Build context: repo root (docker build -f consumer-3/Dockerfile .)

# Stage 1: builder
FROM eclipse-temurin:21 AS builder
WORKDIR /build

# Install Maven (not included in eclipse-temurin:21)
RUN apt-get update -qq && apt-get install -y -qq maven 2>/dev/null

# Copy pom.xml and download deps first (Docker layer cache)
COPY consumer-3/pom.xml .
RUN mvn dependency:go-offline -q

# Copy source
COPY consumer-3/src ./src

# Build fat JAR
RUN mvn package -DskipTests -q

# Stage 2: runtime
FROM eclipse-temurin:21-jre
WORKDIR /app
COPY --from=builder /build/target/consumer3.jar .
COPY producer/calculator-api.json ./calculator-api.json

ENV PRODUCER_URL=http://producer:10001
ENV CVT_SERVER_ADDR=cvt:9550
ENV SCHEMA_PATH=/app/calculator-api.json

ENTRYPOINT ["java", "-jar", "consumer3.jar"]

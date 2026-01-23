# Consumer-1 Contract Tests

This directory contains contract tests demonstrating all three CVT validation approaches for Consumer-1 (Node.js).

## Test Files

| File                   | Approach              | Requires Producer | Requires CVT | Recommended For                |
| ---------------------- | --------------------- | ----------------- | ------------ | ------------------------------ |
| `manual.test.js`       | Manual Validation     | Yes               | Yes          | Full control, custom scenarios |
| `adapter.test.js`      | HTTP Adapter          | Yes               | Yes          | Integration testing            |
| `mock.test.js`         | Mock Client           | No                | Yes          | Unit tests, CI/CD pipelines    |
| `registration.test.js` | Consumer Registration | No                | Yes          | Contract registration          |

## Prerequisites

1. Install dependencies:

   ```bash
   cd consumer-1
   npm install
   ```

2. Start services (for integration tests):

   ```bash
   make up
   ```

## Running Tests

### From consumer-1 directory

```bash
# Run all tests
npm test

# Run mock tests only (no producer needed)
npm run test:mock

# Run integration tests (requires producer)
npm run test:integration

# Run registration tests
npm run test:registration
```

### From project root

```bash
# Run all Consumer-1 tests
make test-consumer-1

# Run mock tests only
make test-consumer-1-mock

# Run integration tests
make test-consumer-1-integration

# Run registration tests
make test-consumer-1-registration
```

## Test Approaches Explained

### Test Setup

Before running tests, the schema is registered with CVT. This happens once at test initialization.

```mermaid
sequenceDiagram
    participant Test
    participant Validator
    participant CVT Server

    Test->>Validator: Initialize with CVT server address
    Validator->>CVT Server: Connect (gRPC)
    Test->>Validator: registerSchema(calculator-api.yaml)
    Validator->>CVT Server: Register OpenAPI schema
    CVT Server-->>Validator: Schema ID
    Note over Test: Ready to run validation tests
```

### 1. Manual Validation (`manual.test.js`)

Makes real HTTP calls to the producer, then explicitly validates the request/response pair against the contract. This approach gives full control over what gets validated and when.

```mermaid
sequenceDiagram
    participant Test
    participant Producer
    participant Validator
    participant CVT Server

    Test->>Producer: GET /add?x=5&y=3
    Producer-->>Test: {result: 8}
    Test->>Test: Build request/response objects
    Test->>Validator: validate(request, response)
    Validator->>CVT Server: Validate against schema (gRPC)
    CVT Server-->>Validator: {valid: true, errors: []}
    Validator-->>Test: Validation result
```

```javascript
const result = await validator.validate(request, response);
expect(result.valid).toBe(true);
```

### 2. HTTP Adapter (`adapter.test.js`)

Automatic validation via axios interceptors. With `autoValidate: true` (the default), every HTTP request is validated transparently.

```mermaid
sequenceDiagram
    participant Test
    participant Axios Interceptor
    participant Producer
    participant Validator
    participant CVT Server

    Test->>Axios Interceptor: createAxiosAdapter(validator)
    Test->>Axios Interceptor: GET /add?x=5&y=3
    Axios Interceptor->>Producer: HTTP Request
    Producer-->>Axios Interceptor: {result: 8}
    Axios Interceptor->>Validator: Auto-validate
    Validator->>CVT Server: Validate (gRPC)
    CVT Server-->>Validator: {valid: true}
    Axios Interceptor->>Axios Interceptor: Store interaction
    Axios Interceptor-->>Test: Response + validation result
```

```javascript
const adapter = createAxiosAdapter({
  axios: client,
  validator,
  autoValidate: true,
});
await client.get("/add", { params: { x: 5, y: 3 } });
// Validation happens automatically
```

### 3. Mock Client (`mock.test.js`)

No real HTTP calls to the producer. Instead, responses are generated directly from the OpenAPI schema. This is ideal for unit testing in isolation or CI/CD pipelines where spinning up the producer isn't practical. The generated responses are schema-compliant (correct structure and types) but won't reflect real business logic. Interactions are still captured and can be used for consumer registration.

```mermaid
sequenceDiagram
    participant Test
    participant Mock Adapter
    participant OpenAPI Schema
    participant Validator
    participant CVT Server

    Test->>Mock Adapter: createMockAdapter(validator)
    Test->>Mock Adapter: fetch("/add?x=5&y=3")
    Mock Adapter->>OpenAPI Schema: Lookup /add response schema
    OpenAPI Schema-->>Mock Adapter: {type: object, properties: {result: number}}
    Mock Adapter->>Mock Adapter: Generate {result: <number>}
    Mock Adapter->>Validator: Validate generated response
    Validator->>CVT Server: Validate (gRPC)
    CVT Server-->>Validator: {valid: true}
    Mock Adapter->>Mock Adapter: Store interaction
    Mock Adapter-->>Test: Generated response
```

```javascript
const response = await mock.fetch("http://calculator-api/add?x=5&y=3");
// Response is generated from schema, no real HTTP call
```

### 4. Consumer Registration (`registration.test.js`)

Registers which endpoints and response fields this consumer depends on. This enables **can-i-deploy checks**: before deploying a new producer version, CVT verifies it won't break existing consumers. If a producer removes or changes a field that a registered consumer depends on, CVT flags it as a breaking change. Interactions captured during testing (via adapters) can be used to auto-generate the consumer registration.

```mermaid
sequenceDiagram
    participant Test
    participant Adapter
    participant Validator
    participant CVT Server

    Note over Test,Adapter: Capture interactions during tests
    Test->>Adapter: Run tests (multiple requests)
    Adapter->>Adapter: Store interactions[]

    Note over Test,CVT Server: Register consumer from interactions
    Test->>Adapter: getInteractions()
    Adapter-->>Test: [{request, response, validation}]
    Test->>Validator: registerConsumerFromInteractions(interactions, config)
    Validator->>CVT Server: Register consumer-1 v1.0.0
    CVT Server-->>Validator: Consumer registered

    Note over Test,CVT Server: Can-I-Deploy check
    Test->>Validator: canIDeploy(schemaId, "1.0.0", "demo")
    Validator->>CVT Server: Check compatibility
    CVT Server-->>Validator: {canDeploy: true}

    Note over Test,CVT Server: Breaking change detection
    Test->>Validator: Register modified schema (field removed)
    Test->>Validator: canIDeploy(newSchemaId, "1.0.0", "demo")
    Validator->>CVT Server: Check compatibility
    CVT Server-->>Validator: {canDeploy: false, reason: "breaking change"}
```

```javascript
// Auto-registration
const consumer = await validator.registerConsumerFromInteractions(interactions, config);

// Manual registration
const consumer = await validator.registerConsumer({ usedEndpoints: [...] });
```

## Endpoints Tested

Consumer-1 uses these Calculator API endpoints:

- `GET /add?x={number}&y={number}` - Addition
- `GET /subtract?x={number}&y={number}` - Subtraction

## Environment Variables

| Variable          | Default                              | Description                           |
| ----------------- | ------------------------------------ | ------------------------------------- |
| `CVT_SERVER_ADDR` | `localhost:9550`                     | CVT server address                    |
| `PRODUCER_URL`    | `http://localhost:10001`             | Producer API URL                      |
| `SCHEMA_PATH`     | `../../producer/calculator-api.yaml` | Path to OpenAPI schema                |
| `CVT_ENVIRONMENT` | `demo`                               | Environment for consumer registration |

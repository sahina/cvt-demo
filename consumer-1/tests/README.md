# Consumer-1 Contract Tests

This directory contains contract tests demonstrating all three CVT validation approaches for Consumer-1 (Node.js).

## Test Files

| File                   | Approach              | Requires Producer |
| ---------------------- | --------------------- | ----------------- |
| `manual.test.js`       | Manual Validation     | Yes               |
| `adapter.test.js`      | HTTP Adapter          | Yes               |
| `mock.test.js`         | Mock Client           | No                |
| `registration.test.js` | Consumer Registration | No                |

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

### 1. Manual Validation (`manual.test.js`)

Tests that explicitly call `validator.validate()` with request/response objects. This approach gives full control over what gets validated.

```javascript
const result = await validator.validate(request, response);
expect(result.valid).toBe(true);
```

### 2. HTTP Adapter (`adapter.test.js`)

Tests that use `createAxiosAdapter()` to automatically validate all HTTP requests made through axios. The adapter intercepts requests and validates them against the schema.

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

Tests that use `createMockAdapter()` to generate schema-compliant responses without a real producer. Useful for unit testing and capturing interactions.

```javascript
const response = await mock.fetch("http://calculator-api/add?x=5&y=3");
// Response is generated from schema, no real HTTP call
```

### 4. Consumer Registration (`registration.test.js`)

Tests demonstrating both auto-registration (from captured interactions) and manual registration (with explicit endpoint definitions).

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

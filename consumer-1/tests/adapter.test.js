/**
 * HTTP ADAPTER APPROACH
 * ---------------------
 * Wraps your HTTP client to automatically validate all requests/responses against the schema.
 *
 * Prerequisites:
 * - CVT server running
 * - Producer service running
 *
 * When to use:
 * - Integration tests against a real or staging API
 * - Catch contract violations automatically during existing tests
 * - Minimal code changes to add validation to existing test suites
 */

const { ContractValidator } = require('@cvt/cvt-sdk');
const { createAxiosAdapter } = require('@cvt/cvt-sdk/adapters');
const axios = require('axios');
const path = require('path');

const CVT_SERVER_ADDR = process.env.CVT_SERVER_ADDR || 'localhost:9550';
const PRODUCER_URL = process.env.PRODUCER_URL || 'http://localhost:10001';
const SCHEMA_PATH = process.env.SCHEMA_PATH || path.join(__dirname, '../../producer/calculator-api.json');

describe('HTTP Adapter Approach', () => {
  let validator;
  let client;
  let adapter;

  beforeAll(async () => {
    validator = new ContractValidator(CVT_SERVER_ADDR);
    await validator.registerSchema('calculator-api', SCHEMA_PATH);

    client = axios.create({
      baseURL: PRODUCER_URL,
      timeout: 5000,
    });

    const validationErrors = [];
    adapter = createAxiosAdapter({
      axios: client,
      validator,
      autoValidate: true,
      onValidationFailure: (result, request, response) => {
        validationErrors.push({ result, request, response });
      },
    });
  });

  afterAll(() => {
    if (adapter) {
      adapter.detach();
    }
    if (validator) {
      validator.close();
    }
  });

  beforeEach(() => {
    adapter.clearInteractions();
  });

  describe('/add endpoint with automatic validation', () => {
    test('should automatically validate add operation', async () => {
      const response = await client.get('/add', { params: { a: 5, b: 3 } });

      expect(response.status).toBe(200);
      expect(response.data.result).toBe(8);

      const interactions = adapter.getInteractions();
      expect(interactions.length).toBe(1);
      expect(interactions[0].validationResult.valid).toBe(true);
    });

    test('should capture request and response details', async () => {
      await client.get('/add', { params: { a: 10, b: 20 } });

      const interactions = adapter.getInteractions();
      expect(interactions.length).toBe(1);

      const interaction = interactions[0];
      expect(interaction.request.method).toBe('GET');
      expect(interaction.request.path).toContain('/add');
      expect(interaction.response.statusCode).toBe(200);
      expect(interaction.response.body.result).toBe(30);
    });
  });

  describe('/subtract endpoint with automatic validation', () => {
    test('should automatically validate subtract operation', async () => {
      const response = await client.get('/subtract', { params: { a: 15, b: 7 } });

      expect(response.status).toBe(200);
      expect(response.data.result).toBe(8);

      const interactions = adapter.getInteractions();
      expect(interactions.length).toBe(1);
      expect(interactions[0].validationResult.valid).toBe(true);
    });

    test('should handle negative result correctly', async () => {
      const response = await client.get('/subtract', { params: { a: 3, b: 10 } });

      expect(response.status).toBe(200);
      expect(response.data.result).toBe(-7);

      const interactions = adapter.getInteractions();
      expect(interactions[0].validationResult.valid).toBe(true);
    });
  });

  describe('multiple operations', () => {
    test('should capture multiple interactions', async () => {
      await client.get('/add', { params: { a: 1, b: 2 } });
      await client.get('/subtract', { params: { a: 5, b: 3 } });
      await client.get('/add', { params: { a: 10, b: 10 } });

      const interactions = adapter.getInteractions();
      expect(interactions.length).toBe(3);
      expect(interactions.every(i => i.validationResult.valid)).toBe(true);
    });
  });
});

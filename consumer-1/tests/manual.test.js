const { ContractValidator } = require('@cvt/cvt-sdk');
const axios = require('axios');
const path = require('path');

const CVT_SERVER_ADDR = process.env.CVT_SERVER_ADDR || 'localhost:9550';
const PRODUCER_URL = process.env.PRODUCER_URL || 'http://localhost:10001';
const SCHEMA_PATH = process.env.SCHEMA_PATH || path.join(__dirname, '../../producer/calculator-api.json');

describe('Manual Validation Approach', () => {
  let validator;
  let client;

  beforeAll(async () => {
    validator = new ContractValidator(CVT_SERVER_ADDR);
    await validator.registerSchema('calculator-api', SCHEMA_PATH);
    client = axios.create({
      baseURL: PRODUCER_URL,
      timeout: 5000,
    });
  });

  afterAll(() => {
    if (validator) {
      validator.close();
    }
  });

  describe('/add endpoint', () => {
    test('should validate successful add operation', async () => {
      const response = await client.get('/add', { params: { a: 5, b: 3 } });

      const request = {
        method: 'GET',
        path: '/add?a=5&b=3',
        headers: {},
      };

      const validationResponse = {
        statusCode: response.status,
        headers: response.headers,
        body: response.data,
      };

      const result = await validator.validate(request, validationResponse);

      expect(result.valid).toBe(true);
      expect(result.errors).toEqual([]);
      expect(response.data.result).toBe(8);
    });

    test('should detect missing result field in response', async () => {
      const request = {
        method: 'GET',
        path: '/add?a=5&b=3',
        headers: {},
      };

      const invalidResponse = {
        statusCode: 200,
        headers: { 'content-type': 'application/json' },
        body: { total: 8 },
      };

      const result = await validator.validate(request, invalidResponse);

      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
    });

    test('should validate 400 error response for invalid input', async () => {
      const request = {
        method: 'GET',
        path: '/add?a=invalid&b=3',
        headers: {},
      };

      const errorResponse = {
        statusCode: 400,
        headers: { 'content-type': 'application/json' },
        body: { error: 'Invalid parameter: a must be a number' },
      };

      const result = await validator.validate(request, errorResponse);

      expect(result.valid).toBe(true);
    });
  });

  describe('/subtract endpoint', () => {
    test('should validate successful subtract operation', async () => {
      const response = await client.get('/subtract', { params: { a: 10, b: 4 } });

      const request = {
        method: 'GET',
        path: '/subtract?a=10&b=4',
        headers: {},
      };

      const validationResponse = {
        statusCode: response.status,
        headers: response.headers,
        body: response.data,
      };

      const result = await validator.validate(request, validationResponse);

      expect(result.valid).toBe(true);
      expect(result.errors).toEqual([]);
      expect(response.data.result).toBe(6);
    });

    test('should detect incorrect response structure', async () => {
      const request = {
        method: 'GET',
        path: '/subtract?a=10&b=4',
        headers: {},
      };

      const invalidResponse = {
        statusCode: 200,
        headers: { 'content-type': 'application/json' },
        body: { difference: 6 },
      };

      const result = await validator.validate(request, invalidResponse);

      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
    });
  });
});

const { ContractValidator } = require('@cvt/cvt-sdk');
const { createMockAdapter } = require('@cvt/cvt-sdk/adapters');
const path = require('path');

const CVT_SERVER_ADDR = process.env.CVT_SERVER_ADDR || 'localhost:9550';
const SCHEMA_PATH = process.env.SCHEMA_PATH || path.join(__dirname, '../../producer/calculator-api.json');

describe('Mock Client Approach', () => {
  let validator;
  let mock;

  beforeAll(async () => {
    validator = new ContractValidator(CVT_SERVER_ADDR);
    await validator.registerSchema('calculator-api', SCHEMA_PATH);

    mock = createMockAdapter({
      validator,
      cache: true,
      generateOptions: { useExamples: true },
    });
  });

  afterAll(() => {
    if (mock) {
      mock.clearCache();
    }
    if (validator) {
      validator.close();
    }
  });

  beforeEach(() => {
    mock.clearInteractions();
  });

  describe('/add endpoint mocking', () => {
    test('should generate valid mock response for add', async () => {
      const response = await mock.fetch('http://calculator-api/add?a=5&b=3');
      const data = await response.json();

      expect(response.status).toBe(200);
      expect(data).toHaveProperty('result');
      expect(typeof data.result).toBe('number');
    });

    test('should capture mock interactions', async () => {
      await mock.fetch('http://calculator-api/add?a=10&b=20');

      const interactions = mock.getInteractions();
      expect(interactions.length).toBe(1);
      expect(interactions[0].request.method).toBe('GET');
      expect(interactions[0].request.path).toContain('/add');
    });

    test('should generate response matching schema', async () => {
      const response = await mock.fetch('http://calculator-api/add?a=1&b=1');
      const data = await response.json();

      const validationResult = await validator.validate(
        { method: 'GET', path: '/add?a=1&b=1', headers: {} },
        { statusCode: 200, headers: { 'content-type': 'application/json' }, body: data }
      );

      expect(validationResult.valid).toBe(true);
    });
  });

  describe('/subtract endpoint mocking', () => {
    test('should generate valid mock response for subtract', async () => {
      const response = await mock.fetch('http://calculator-api/subtract?a=10&b=4');
      const data = await response.json();

      expect(response.status).toBe(200);
      expect(data).toHaveProperty('result');
      expect(typeof data.result).toBe('number');
    });

    test('should capture subtract interaction', async () => {
      await mock.fetch('http://calculator-api/subtract?a=100&b=50');

      const interactions = mock.getInteractions();
      expect(interactions.length).toBe(1);
      expect(interactions[0].request.path).toContain('/subtract');
    });
  });

  describe('interaction capture for registration', () => {
    test('should capture all Consumer-1 endpoints for registration', async () => {
      await mock.fetch('http://calculator-api/add?a=5&b=3');
      await mock.fetch('http://calculator-api/subtract?a=10&b=4');

      const interactions = mock.getInteractions();
      expect(interactions.length).toBe(2);

      const paths = interactions.map(i => i.request.path);
      expect(paths.some(p => p.includes('/add'))).toBe(true);
      expect(paths.some(p => p.includes('/subtract'))).toBe(true);
    });
  });
});

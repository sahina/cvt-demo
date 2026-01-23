/**
 * CONSUMER REGISTRATION
 * ---------------------
 * Registers which API endpoints your service uses so breaking changes can be detected.
 *
 * Prerequisites:
 * - CVT server running
 *
 * Key concepts:
 * - Auto-registration: Capture interactions from mocks/adapters, then register
 * - Manual registration: Explicitly declare used endpoints
 * - can-i-deploy: Check if a schema change is safe for all consumers
 */

const { ContractValidator } = require('@cvt/cvt-sdk');
const { createMockAdapter } = require('@cvt/cvt-sdk/adapters');
const path = require('path');

const CVT_SERVER_ADDR = process.env.CVT_SERVER_ADDR || 'localhost:9550';
const SCHEMA_PATH = process.env.SCHEMA_PATH || path.join(__dirname, '../../producer/calculator-api.json');
const CONSUMER_ID = 'consumer-1';
const CONSUMER_VERSION = process.env.npm_package_version || '1.0.0';
const ENVIRONMENT = process.env.CVT_ENVIRONMENT || 'demo';

describe('Consumer Registration', () => {
  let validator;
  let mock;

  beforeAll(async () => {
    validator = new ContractValidator(CVT_SERVER_ADDR);
    await validator.registerSchema('calculator-api', SCHEMA_PATH);

    mock = createMockAdapter({
      validator,
      cache: true,
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

  describe('Auto-registration from interactions', () => {
    test('should capture interactions for auto-registration', async () => {
      await mock.fetch('http://calculator-api/add?a=5&b=3');
      await mock.fetch('http://calculator-api/subtract?a=10&b=4');

      const interactions = mock.getInteractions();
      expect(interactions.length).toBe(2);

      const opts = validator.buildConsumerFromInteractions(interactions, {
        consumerId: CONSUMER_ID,
        consumerVersion: CONSUMER_VERSION,
        environment: ENVIRONMENT,
        schemaVersion: '1.0.0',
        schemaId: 'calculator-api',
      });

      expect(opts.consumerId).toBe(CONSUMER_ID);
      expect(opts.usedEndpoints).toBeDefined();
      expect(opts.usedEndpoints.length).toBe(2);

      const endpoints = opts.usedEndpoints.map(e => `${e.method} ${e.path}`);
      expect(endpoints).toContain('GET /add');
      expect(endpoints).toContain('GET /subtract');
    });

    test('should register consumer from captured interactions', async () => {
      await mock.fetch('http://calculator-api/add?a=1&b=2');
      await mock.fetch('http://calculator-api/subtract?a=5&b=3');

      const interactions = mock.getInteractions();

      const consumer = await validator.registerConsumerFromInteractions(
        interactions,
        {
          consumerId: CONSUMER_ID,
          consumerVersion: CONSUMER_VERSION,
          environment: ENVIRONMENT,
          schemaVersion: '1.0.0',
          schemaId: 'calculator-api',
        }
      );

      expect(consumer).toBeDefined();
      expect(consumer.consumerId).toBe(CONSUMER_ID);
    });
  });

  describe('Manual registration', () => {
    test('should manually register consumer with explicit endpoints', async () => {
      const consumer = await validator.registerConsumer({
        consumerId: CONSUMER_ID,
        consumerVersion: CONSUMER_VERSION,
        schemaId: 'calculator-api',
        schemaVersion: '1.0.0',
        environment: ENVIRONMENT,
        usedEndpoints: [
          { method: 'GET', path: '/add', usedFields: ['result'] },
          { method: 'GET', path: '/subtract', usedFields: ['result'] },
        ],
      });

      expect(consumer).toBeDefined();
      expect(consumer.consumerId).toBe(CONSUMER_ID);
    });

    test('should list registered consumers', async () => {
      await validator.registerConsumer({
        consumerId: CONSUMER_ID,
        consumerVersion: CONSUMER_VERSION,
        schemaId: 'calculator-api',
        schemaVersion: '1.0.0',
        environment: ENVIRONMENT,
        usedEndpoints: [
          { method: 'GET', path: '/add', usedFields: ['result'] },
          { method: 'GET', path: '/subtract', usedFields: ['result'] },
        ],
      });

      const consumers = await validator.listConsumers('calculator-api', ENVIRONMENT);

      expect(consumers).toBeDefined();
      expect(Array.isArray(consumers)).toBe(true);
      expect(consumers.some(c => c.consumerId === CONSUMER_ID)).toBe(true);
    });
  });

  describe('Breaking change detection', () => {
    test('should check deployment safety with can-i-deploy', async () => {
      await validator.registerConsumer({
        consumerId: CONSUMER_ID,
        consumerVersion: CONSUMER_VERSION,
        schemaId: 'calculator-api',
        schemaVersion: '1.0.0',
        environment: ENVIRONMENT,
        usedEndpoints: [
          { method: 'GET', path: '/add', usedFields: ['result'] },
          { method: 'GET', path: '/subtract', usedFields: ['result'] },
        ],
      });

      const result = await validator.canIDeploy('calculator-api', '1.0.0', ENVIRONMENT);

      expect(result).toBeDefined();
      expect(result.safeToDeploy).toBeDefined();
    });
  });
});

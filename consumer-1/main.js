#!/usr/bin/env node
/**
 * Consumer-1: A CLI tool that uses the Calculator API for add and subtract operations.
 *
 * Usage:
 *   node main.js add <a> <b> [--validate]
 *   node main.js subtract <a> <b> [--validate]
 *
 * Options:
 *   --validate  Enable CVT contract validation (default: off)
 */

const { program } = require('commander');
const axios = require('axios');

// Configuration
const PRODUCER_URL = process.env.PRODUCER_URL || 'http://localhost:10001';
const CVT_SERVER_ADDR = process.env.CVT_SERVER_ADDR || 'localhost:9550';
const SCHEMA_PATH = process.env.SCHEMA_PATH || './calculator-api.yaml';

/**
 * Creates an axios instance, optionally wrapped with CVT validation.
 */
async function createClient(validate) {
  const client = axios.create({
    baseURL: PRODUCER_URL,
    timeout: 5000,
    headers: {
      'Accept': 'application/json',
    },
  });

  if (validate) {
    try {
      const { ContractValidator } = require('@cvt/cvt-sdk');

      const validator = new ContractValidator(CVT_SERVER_ADDR);
      await validator.registerSchema('calculator-api', SCHEMA_PATH);

      // Add response interceptor for manual validation
      client.interceptors.response.use(
        async (response) => {
          // Build path with query string from params
          let path = response.config.url;
          if (response.config.params) {
            const queryString = new URLSearchParams(response.config.params).toString();
            path = `${path}?${queryString}`;
          }

          const request = {
            method: response.config.method.toUpperCase(),
            path: path,
            headers: response.config.headers,
          };

          const validationResponse = {
            statusCode: response.status,
            headers: response.headers,
            body: response.data,
          };

          const result = await validator.validate(request, validationResponse);

          if (!result.valid) {
            console.error('CVT Validation failed:', result.errors);
            process.exit(1);
          }

          return response;
        },
        (error) => Promise.reject(error)
      );

      console.log('CVT validation enabled');
    } catch (err) {
      console.error('Warning: Failed to enable CVT validation:', err.message);
      console.log('Continuing without validation...');
    }
  }

  return client;
}

/**
 * Performs an add operation.
 */
async function add(a, b, options) {
  try {
    const client = await createClient(options.validate);
    const response = await client.get('/add', {
      params: { a, b },
    });

    console.log(`${a} + ${b} = ${response.data.result}`);
  } catch (error) {
    handleError(error);
  }
}

/**
 * Performs a subtract operation.
 */
async function subtract(a, b, options) {
  try {
    const client = await createClient(options.validate);
    const response = await client.get('/subtract', {
      params: { a, b },
    });

    console.log(`${a} - ${b} = ${response.data.result}`);
  } catch (error) {
    handleError(error);
  }
}

/**
 * Handles errors from API calls.
 */
function handleError(error) {
  if (error.response) {
    // The request was made and the server responded with a status code
    // that falls out of the range of 2xx
    if (error.response.data && error.response.data.error) {
      console.error('Error:', error.response.data.error);
    } else {
      console.error('Error:', error.response.status, error.response.statusText);
    }
  } else if (error.request) {
    // The request was made but no response was received
    console.error('Error: No response from server. Is the producer running?');
  } else {
    // Something happened in setting up the request
    console.error('Error:', error.message);
  }
  process.exit(1);
}

// CLI Setup
program
  .name('consumer-1')
  .description('CLI tool for add and subtract operations using the Calculator API')
  .version('1.0.0');

program
  .command('add <a> <b>')
  .description('Add two numbers')
  .option('--validate', 'Enable CVT contract validation', false)
  .action((a, b, options) => {
    const numA = parseFloat(a);
    const numB = parseFloat(b);

    if (isNaN(numA) || isNaN(numB)) {
      console.error('Error: Both arguments must be valid numbers');
      process.exit(1);
    }

    add(numA, numB, options);
  });

program
  .command('subtract <a> <b>')
  .description('Subtract two numbers')
  .option('--validate', 'Enable CVT contract validation', false)
  .action((a, b, options) => {
    const numA = parseFloat(a);
    const numB = parseFloat(b);

    if (isNaN(numA) || isNaN(numB)) {
      console.error('Error: Both arguments must be valid numbers');
      process.exit(1);
    }

    subtract(numA, numB, options);
  });

program.parse();

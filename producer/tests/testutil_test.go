// Package tests provides test utilities and shared test code for producer testing.
package tests

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sahina/cvt/sdks/go/cvt"
	"github.com/sahina/cvt/sdks/go/cvt/producer"
)

// TestConfig holds configuration for tests.
type TestConfig struct {
	CVTServerAddr string
	ProducerURL   string
	SchemaPath    string
	SchemaID      string
	Environment   string
}

// GetTestConfig returns test configuration from environment variables.
func GetTestConfig(t *testing.T) *TestConfig {
	t.Helper()

	cvtServerAddr := os.Getenv("CVT_SERVER_ADDR")
	if cvtServerAddr == "" {
		cvtServerAddr = "localhost:9550"
	}

	producerURL := os.Getenv("PRODUCER_URL")
	if producerURL == "" {
		producerURL = "http://localhost:10001"
	}

	// Get the schema path relative to the test directory
	schemaPath := os.Getenv("SCHEMA_PATH")
	if schemaPath == "" {
		// Default to the JSON schema in the producer directory (CVT SDK requires JSON)
		schemaPath = filepath.Join("..", "calculator-api.json")
	}

	environment := os.Getenv("CVT_ENVIRONMENT")
	if environment == "" {
		environment = "demo"
	}

	return &TestConfig{
		CVTServerAddr: cvtServerAddr,
		ProducerURL:   producerURL,
		SchemaPath:    schemaPath,
		SchemaID:      "calculator-api",
		Environment:   environment,
	}
}

// NewTestValidator creates a new CVT validator for testing.
func NewTestValidator(t *testing.T, config *TestConfig) *cvt.Validator {
	t.Helper()

	validator, err := cvt.NewValidator(config.CVTServerAddr)
	if err != nil {
		t.Fatalf("Failed to create CVT validator: %v", err)
	}

	ctx := context.Background()
	if err := validator.RegisterSchema(ctx, config.SchemaID, config.SchemaPath); err != nil {
		validator.Close()
		t.Fatalf("Failed to register schema: %v", err)
	}

	return validator
}

// NewProducerTestKit creates a new ProducerTestKit for testing.
// Note: Schema must be registered with CVT server before calling this.
// Use NewTestValidator to register the schema first.
func NewProducerTestKit(t *testing.T, config *TestConfig) *producer.ProducerTestKit {
	t.Helper()

	// First, ensure the schema is registered
	validator := NewTestValidator(t, config)
	validator.Close() // Close the validator, schema remains registered

	testKit, err := producer.NewProducerTestKit(producer.TestConfig{
		SchemaID:      config.SchemaID,
		ServerAddress: config.CVTServerAddr,
	})
	if err != nil {
		t.Fatalf("Failed to create ProducerTestKit: %v", err)
	}

	return testKit
}

// CalculatorTestCase represents a test case for calculator operations.
type CalculatorTestCase struct {
	Name           string
	Path           string
	Method         string
	X              float64
	Y              float64
	ExpectedResult float64
	ExpectedStatus int
	ExpectError    bool
	ErrorContains  string
}

// CalculatorTestCases returns standard test cases for all calculator operations.
func CalculatorTestCases() []CalculatorTestCase {
	return []CalculatorTestCase{
		// Add tests
		{
			Name:           "add positive numbers",
			Path:           "/add",
			Method:         "GET",
			X:              5,
			Y:              3,
			ExpectedResult: 8,
			ExpectedStatus: 200,
		},
		{
			Name:           "add negative numbers",
			Path:           "/add",
			Method:         "GET",
			X:              -5,
			Y:              -3,
			ExpectedResult: -8,
			ExpectedStatus: 200,
		},
		{
			Name:           "add with zero",
			Path:           "/add",
			Method:         "GET",
			X:              10,
			Y:              0,
			ExpectedResult: 10,
			ExpectedStatus: 200,
		},
		// Subtract tests
		{
			Name:           "subtract positive numbers",
			Path:           "/subtract",
			Method:         "GET",
			X:              10,
			Y:              4,
			ExpectedResult: 6,
			ExpectedStatus: 200,
		},
		{
			Name:           "subtract resulting in negative",
			Path:           "/subtract",
			Method:         "GET",
			X:              3,
			Y:              7,
			ExpectedResult: -4,
			ExpectedStatus: 200,
		},
		// Multiply tests
		{
			Name:           "multiply positive numbers",
			Path:           "/multiply",
			Method:         "GET",
			X:              4,
			Y:              7,
			ExpectedResult: 28,
			ExpectedStatus: 200,
		},
		{
			Name:           "multiply with zero",
			Path:           "/multiply",
			Method:         "GET",
			X:              100,
			Y:              0,
			ExpectedResult: 0,
			ExpectedStatus: 200,
		},
		{
			Name:           "multiply negative numbers",
			Path:           "/multiply",
			Method:         "GET",
			X:              -3,
			Y:              -4,
			ExpectedResult: 12,
			ExpectedStatus: 200,
		},
		// Divide tests
		{
			Name:           "divide positive numbers",
			Path:           "/divide",
			Method:         "GET",
			X:              10,
			Y:              2,
			ExpectedResult: 5,
			ExpectedStatus: 200,
		},
		{
			Name:           "divide with decimal result",
			Path:           "/divide",
			Method:         "GET",
			X:              7,
			Y:              2,
			ExpectedResult: 3.5,
			ExpectedStatus: 200,
		},
		// Division by zero error case
		{
			Name:           "divide by zero",
			Path:           "/divide",
			Method:         "GET",
			X:              10,
			Y:              0,
			ExpectedStatus: 400,
			ExpectError:    true,
			ErrorContains:  "division by zero",
		},
	}
}

// ErrorTestCases returns test cases that should produce errors.
func ErrorTestCases() []CalculatorTestCase {
	return []CalculatorTestCase{
		{
			Name:           "missing parameters",
			Path:           "/add",
			Method:         "GET",
			ExpectedStatus: 400,
			ExpectError:    true,
			ErrorContains:  "missing required parameters",
		},
		{
			Name:           "invalid x parameter",
			Path:           "/add",
			Method:         "GET",
			ExpectedStatus: 400,
			ExpectError:    true,
			ErrorContains:  "must be a valid number",
		},
	}
}

// ValidatorAdapter adapts cvt.Validator to producer.Validator interface.
// This is used for middleware testing.
type ValidatorAdapter struct {
	Validator *cvt.Validator
}

// Validate implements the producer.Validator interface.
func (a *ValidatorAdapter) Validate(ctx context.Context, schemaID string, interaction *producer.Interaction) (*producer.ValidationResult, error) {
	// Convert producer.Interaction to cvt.ValidationRequest and cvt.ValidationResponse
	request := cvt.ValidationRequest{
		Method:  interaction.Method,
		Path:    interaction.Path,
		Headers: interaction.Headers,
	}

	// Parse request body if present
	if interaction.Body != "" {
		var body any
		if err := json.Unmarshal([]byte(interaction.Body), &body); err == nil {
			request.Body = body
		} else {
			request.Body = interaction.Body
		}
	}

	response := cvt.ValidationResponse{
		StatusCode: interaction.StatusCode,
		Headers:    interaction.ResponseHeaders,
	}

	// Parse response body if present
	if interaction.ResponseBody != "" {
		var body any
		if err := json.Unmarshal([]byte(interaction.ResponseBody), &body); err == nil {
			response.Body = body
		} else {
			response.Body = interaction.ResponseBody
		}
	}

	// Call the underlying validator
	result, err := a.Validator.Validate(ctx, request, response)
	if err != nil {
		return nil, err
	}

	return &producer.ValidationResult{
		Valid:  result.Valid,
		Errors: result.Errors,
	}, nil
}

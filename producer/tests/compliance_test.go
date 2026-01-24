// Package tests contains producer contract tests.
package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sahina/cvt-demo/producer/handlers"
	"github.com/sahina/cvt/sdks/go/cvt/producer"
)

// httpHeaderToMap converts http.Header to map[string]string for CVT SDK.
func httpHeaderToMap(h http.Header) map[string]string {
	result := make(map[string]string)
	for k, v := range h {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

// parseBody parses JSON body bytes into a map for CVT validation.
func parseBody(body []byte) any {
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return string(body)
	}
	return result
}

// TestSchemaCompliance_AllOperations tests that all calculator operations
// produce responses that comply with the OpenAPI schema.
func TestSchemaCompliance_AllOperations(t *testing.T) {
	config := GetTestConfig(t)
	testKit := NewProducerTestKit(t, config)
	defer testKit.Close()

	calc := handlers.NewCalculator()
	ctx := context.Background()

	testCases := CalculatorTestCases()

	for _, tc := range testCases {
		if tc.ExpectError {
			continue // Skip error cases, handled in separate test
		}

		t.Run(tc.Name, func(t *testing.T) {
			// Build the request path with query parameters
			path := fmt.Sprintf("%s?x=%v&y=%v", tc.Path, tc.X, tc.Y)

			// Create HTTP test request
			req := httptest.NewRequest(tc.Method, path, nil)
			rec := httptest.NewRecorder()

			// Call the appropriate handler
			switch tc.Path {
			case "/add":
				calc.Add(rec, req)
			case "/subtract":
				calc.Subtract(rec, req)
			case "/multiply":
				calc.Multiply(rec, req)
			case "/divide":
				calc.Divide(rec, req)
			}

			// Verify status code
			if rec.Code != tc.ExpectedStatus {
				t.Errorf("Expected status %d, got %d", tc.ExpectedStatus, rec.Code)
			}

			// Validate response against schema using ProducerTestKit
			result, err := testKit.ValidateResponse(ctx, producer.ValidateResponseParams{
				Method: tc.Method,
				Path:   path,
				Response: producer.TestResponseData{
					StatusCode: rec.Code,
					Body:       parseBody(rec.Body.Bytes()),
					Headers:    httpHeaderToMap(rec.Header()),
				},
			})

			if err != nil {
				t.Fatalf("Validation error: %v", err)
			}

			if !result.Valid {
				t.Errorf("Response does not comply with schema: %v", result.Errors)
			}
		})
	}
}

// TestSchemaCompliance_NegativeTests tests that CVT correctly identifies
// schema violations in producer responses.
func TestSchemaCompliance_NegativeTests(t *testing.T) {
	config := GetTestConfig(t)
	testKit := NewProducerTestKit(t, config)
	defer testKit.Close()

	ctx := context.Background()

	negativeTests := []struct {
		name          string
		path          string
		method        string
		statusCode    int
		body          []byte
		expectInvalid bool
		description   string
	}{
		{
			name:          "wrong field name (value instead of result)",
			path:          "/add?x=5&y=3",
			method:        "GET",
			statusCode:    200,
			body:          []byte(`{"value": 8}`),
			expectInvalid: true,
			description:   "Schema requires 'result' field, not 'value'",
		},
		{
			name:          "wrong type (string instead of number)",
			path:          "/add?x=5&y=3",
			method:        "GET",
			statusCode:    200,
			body:          []byte(`{"result": "8"}`),
			expectInvalid: true,
			description:   "Schema requires 'result' to be a number, not string",
		},
		{
			name:          "missing required field",
			path:          "/add?x=5&y=3",
			method:        "GET",
			statusCode:    200,
			body:          []byte(`{}`),
			expectInvalid: true,
			description:   "Schema requires 'result' field to be present",
		},
		{
			name:          "extra field allowed (additionalProperties default)",
			path:          "/add?x=5&y=3",
			method:        "GET",
			statusCode:    200,
			body:          []byte(`{"result": 8, "extra": "field"}`),
			expectInvalid: false,
			description:   "Extra fields may be allowed depending on schema settings",
		},
	}

	for _, tc := range negativeTests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := testKit.ValidateResponse(ctx, producer.ValidateResponseParams{
				Method: tc.method,
				Path:   tc.path,
				Response: producer.TestResponseData{
					StatusCode: tc.statusCode,
					Body:       parseBody(tc.body),
					Headers:    map[string]string{"Content-Type": "application/json"},
				},
			})

			if err != nil {
				t.Fatalf("Validation error: %v", err)
			}

			if tc.expectInvalid && result.Valid {
				t.Errorf("Expected validation to fail (%s), but it passed", tc.description)
			}

			if !tc.expectInvalid && !result.Valid {
				t.Errorf("Expected validation to pass (%s), but it failed: %v", tc.description, result.Errors)
			}
		})
	}
}

// TestSchemaCompliance_ErrorResponses tests that error responses (400 status)
// also comply with the error schema.
func TestSchemaCompliance_ErrorResponses(t *testing.T) {
	config := GetTestConfig(t)
	testKit := NewProducerTestKit(t, config)
	defer testKit.Close()

	calc := handlers.NewCalculator()
	ctx := context.Background()

	errorCases := []struct {
		name       string
		path       string
		queryPath  string
		handler    func(http.ResponseWriter, *http.Request)
		statusCode int
	}{
		{
			name:       "division by zero error",
			path:       "/divide",
			queryPath:  "/divide?x=10&y=0",
			handler:    calc.Divide,
			statusCode: 400,
		},
		{
			name:       "missing parameters error",
			path:       "/add",
			queryPath:  "/add",
			handler:    calc.Add,
			statusCode: 400,
		},
		{
			name:       "missing y parameter",
			path:       "/multiply",
			queryPath:  "/multiply?x=5",
			handler:    calc.Multiply,
			statusCode: 400,
		},
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.queryPath, nil)
			rec := httptest.NewRecorder()

			tc.handler(rec, req)

			// Verify we got an error status
			if rec.Code != tc.statusCode {
				t.Errorf("Expected status %d, got %d", tc.statusCode, rec.Code)
			}

			// Validate error response against schema
			result, err := testKit.ValidateResponse(ctx, producer.ValidateResponseParams{
				Method: "GET",
				Path:   tc.queryPath,
				Response: producer.TestResponseData{
					StatusCode: rec.Code,
					Body:       parseBody(rec.Body.Bytes()),
					Headers:    httpHeaderToMap(rec.Header()),
				},
			})

			if err != nil {
				t.Fatalf("Validation error: %v", err)
			}

			if !result.Valid {
				t.Errorf("Error response does not comply with schema: %v", result.Errors)
			}
		})
	}
}

// TestSchemaCompliance_HealthEndpoint tests the health endpoint response.
func TestSchemaCompliance_HealthEndpoint(t *testing.T) {
	calc := handlers.NewCalculator()

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	calc.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Health endpoint is excluded from CVT validation in production,
	// but we can verify the response structure
	body := rec.Body.String()
	if body != `{"status":"healthy"}`+"\n" {
		t.Errorf("Unexpected health response: %s", body)
	}
}

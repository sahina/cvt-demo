// Package tests contains producer contract tests.
package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/sahina/cvt-demo/producer/handlers"
	"github.com/sahina/cvt/sdks/go/cvt/producer"
)

// respHeaderToMap converts http.Header from response to map[string]string for CVT SDK.
func respHeaderToMap(h http.Header) map[string]string {
	result := make(map[string]string)
	for k, v := range h {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

// parseResponseBody parses JSON body bytes into a map for CVT validation.
func parseResponseBody(body []byte) any {
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return string(body)
	}
	return result
}

// TestIntegration_AllEndpoints tests all calculator endpoints via HTTP
// against the running producer service.
func TestIntegration_AllEndpoints(t *testing.T) {
	config := GetTestConfig(t)

	testCases := CalculatorTestCases()

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Build the full URL
			url := fmt.Sprintf("%s%s?x=%v&y=%v", config.ProducerURL, tc.Path, tc.X, tc.Y)

			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("HTTP request failed: %v", err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tc.ExpectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d, got %d. Body: %s", tc.ExpectedStatus, resp.StatusCode, body)
				return
			}

			// Parse response
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response: %v", err)
			}

			if tc.ExpectError {
				var errResp handlers.ErrorResponse
				if err := json.Unmarshal(body, &errResp); err != nil {
					t.Fatalf("Failed to parse error response: %v", err)
				}

				if tc.ErrorContains != "" && errResp.Error == "" {
					t.Errorf("Expected error containing '%s', got empty error", tc.ErrorContains)
				}
			} else {
				var result handlers.ResultResponse
				if err := json.Unmarshal(body, &result); err != nil {
					t.Fatalf("Failed to parse result response: %v", err)
				}

				if result.Result != tc.ExpectedResult {
					t.Errorf("Expected result %v, got %v", tc.ExpectedResult, result.Result)
				}
			}
		})
	}
}

// TestIntegration_HealthEndpoint tests the health endpoint.
func TestIntegration_HealthEndpoint(t *testing.T) {
	config := GetTestConfig(t)

	url := fmt.Sprintf("%s/health", config.ProducerURL)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var health handlers.HealthResponse
	if err := json.Unmarshal(body, &health); err != nil {
		t.Fatalf("Failed to parse health response: %v", err)
	}

	if health.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", health.Status)
	}
}

// TestIntegration_WithCVTValidation tests HTTP responses and validates them
// against the schema using ProducerTestKit.
func TestIntegration_WithCVTValidation(t *testing.T) {
	config := GetTestConfig(t)
	testKit := NewProducerTestKit(t, config)
	defer testKit.Close()

	ctx := context.Background()

	testCases := []struct {
		name           string
		path           string
		x              float64
		y              float64
		expectedStatus int
	}{
		{"add", "/add", 5, 3, 200},
		{"subtract", "/subtract", 10, 4, 200},
		{"multiply", "/multiply", 4, 7, 200},
		{"divide", "/divide", 10, 2, 200},
		{"divide by zero", "/divide", 10, 0, 400},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Make HTTP request to producer
			url := fmt.Sprintf("%s%s?x=%v&y=%v", config.ProducerURL, tc.path, tc.x, tc.y)
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("HTTP request failed: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response: %v", err)
			}

			// Check status code
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, resp.StatusCode)
			}

			// Validate response against schema
			path := fmt.Sprintf("%s?x=%v&y=%v", tc.path, tc.x, tc.y)
			result, err := testKit.ValidateResponse(ctx, producer.ValidateResponseParams{
				Method: "GET",
				Path:   path,
				Response: producer.TestResponseData{
					StatusCode: resp.StatusCode,
					Body:       parseResponseBody(body),
					Headers:    respHeaderToMap(resp.Header),
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

// TestIntegration_ErrorResponses tests that error responses from the producer
// comply with the schema.
func TestIntegration_ErrorResponses(t *testing.T) {
	config := GetTestConfig(t)
	testKit := NewProducerTestKit(t, config)
	defer testKit.Close()

	ctx := context.Background()

	errorCases := []struct {
		name           string
		url            string
		path           string
		expectedStatus int
	}{
		{
			name:           "division by zero",
			url:            "/divide?x=10&y=0",
			path:           "/divide?x=10&y=0",
			expectedStatus: 400,
		},
		{
			name:           "missing parameters",
			url:            "/add",
			path:           "/add",
			expectedStatus: 400,
		},
		{
			name:           "missing y parameter",
			url:            "/multiply?x=5",
			path:           "/multiply?x=5",
			expectedStatus: 400,
		},
		{
			name:           "invalid x parameter",
			url:            "/add?x=abc&y=3",
			path:           "/add?x=abc&y=3",
			expectedStatus: 400,
		},
	}

	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			url := fmt.Sprintf("%s%s", config.ProducerURL, tc.url)
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("HTTP request failed: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response: %v", err)
			}

			// Check we got an error status
			if resp.StatusCode != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tc.expectedStatus, resp.StatusCode, body)
				return
			}

			// Validate error response against schema
			result, err := testKit.ValidateResponse(ctx, producer.ValidateResponseParams{
				Method: "GET",
				Path:   tc.path,
				Response: producer.TestResponseData{
					StatusCode: resp.StatusCode,
					Body:       parseResponseBody(body),
					Headers:    respHeaderToMap(resp.Header),
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

// TestIntegration_ConcurrentRequests tests that the producer handles
// concurrent requests correctly.
func TestIntegration_ConcurrentRequests(t *testing.T) {
	config := GetTestConfig(t)

	// Number of concurrent requests per endpoint
	const concurrency = 10

	endpoints := []struct {
		path           string
		x, y           float64
		expectedResult float64
	}{
		{"/add", 5, 3, 8},
		{"/subtract", 10, 4, 6},
		{"/multiply", 4, 7, 28},
		{"/divide", 10, 2, 5},
	}

	for _, ep := range endpoints {
		t.Run(ep.path, func(t *testing.T) {
			results := make(chan error, concurrency)

			for i := 0; i < concurrency; i++ {
				go func() {
					url := fmt.Sprintf("%s%s?x=%v&y=%v", config.ProducerURL, ep.path, ep.x, ep.y)
					resp, err := http.Get(url)
					if err != nil {
						results <- err
						return
					}
					defer resp.Body.Close()

					if resp.StatusCode != http.StatusOK {
						results <- fmt.Errorf("expected status 200, got %d", resp.StatusCode)
						return
					}

					body, err := io.ReadAll(resp.Body)
					if err != nil {
						results <- err
						return
					}

					var result handlers.ResultResponse
					if err := json.Unmarshal(body, &result); err != nil {
						results <- err
						return
					}

					if result.Result != ep.expectedResult {
						results <- fmt.Errorf("expected %v, got %v", ep.expectedResult, result.Result)
						return
					}

					results <- nil
				}()
			}

			// Collect results
			for i := 0; i < concurrency; i++ {
				if err := <-results; err != nil {
					t.Errorf("Concurrent request %d failed: %v", i, err)
				}
			}
		})
	}
}

// TestIntegration_ContentType tests that responses have correct Content-Type.
func TestIntegration_ContentType(t *testing.T) {
	config := GetTestConfig(t)

	endpoints := []string{"/add?x=1&y=1", "/subtract?x=1&y=1", "/multiply?x=1&y=1", "/divide?x=1&y=1", "/health"}

	for _, ep := range endpoints {
		t.Run(ep, func(t *testing.T) {
			url := fmt.Sprintf("%s%s", config.ProducerURL, ep)
			resp, err := http.Get(url)
			if err != nil {
				t.Fatalf("HTTP request failed: %v", err)
			}
			defer resp.Body.Close()

			contentType := resp.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
			}
		})
	}
}

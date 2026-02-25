package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sahina/cvt/sdks/go/cvt"
)

// TestManual_AddValidation validates a real add response manually
func TestManual_AddValidation(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()

	url := fmt.Sprintf("%s/add?x=4&y=7", config.ProducerURL)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	result, err := validator.Validate(ctx,
		cvt.ValidationRequest{
			Method:  "GET",
			Path:    "/add?x=4&y=7",
			Headers: map[string]string{},
		},
		cvt.ValidationResponse{
			StatusCode: resp.StatusCode,
			Headers:    headersToMap(resp.Header),
			Body:       body,
		},
	)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}
	if !result.Valid {
		t.Errorf("Expected valid response: %v", result.Errors)
	}

	if val, ok := body["result"].(float64); !ok || val != 11 {
		t.Errorf("Expected result=11, got: %v", body["result"])
	}
}

// TestManual_SubtractValidation validates a real subtract response manually
func TestManual_SubtractValidation(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()

	url := fmt.Sprintf("%s/subtract?x=10&y=3", config.ProducerURL)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	result, err := validator.Validate(ctx,
		cvt.ValidationRequest{Method: "GET", Path: "/subtract?x=10&y=3"},
		cvt.ValidationResponse{StatusCode: resp.StatusCode, Body: body},
	)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}
	if !result.Valid {
		t.Errorf("Expected valid: %v", result.Errors)
	}

	if val, ok := body["result"].(float64); !ok || val != 7 {
		t.Errorf("Expected result=7, got: %v", body["result"])
	}
}

// TestManual_DetectsInvalidResponse verifies schema violation is caught
func TestManual_DetectsInvalidResponse(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()

	result, err := validator.Validate(ctx,
		cvt.ValidationRequest{Method: "GET", Path: "/subtract?x=10&y=3"},
		cvt.ValidationResponse{
			StatusCode: 200,
			Body:       map[string]interface{}{"difference": 7},
		},
	)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}
	if result.Valid {
		t.Error("Expected validation to fail for response with 'difference' instead of 'result'")
	}
}

// TestManual_SubtractByNegativeValidation validates subtraction resulting in negative number
func TestManual_SubtractByNegativeValidation(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()

	url := fmt.Sprintf("%s/subtract?x=5&y=10", config.ProducerURL)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	result, err := validator.Validate(ctx,
		cvt.ValidationRequest{Method: "GET", Path: "/subtract?x=5&y=10"},
		cvt.ValidationResponse{StatusCode: resp.StatusCode, Body: body},
	)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}
	if !result.Valid {
		t.Errorf("Negative subtract result should be valid: %v", result.Errors)
	}

	if val, ok := body["result"].(float64); !ok || val != -5 {
		t.Errorf("Expected result=-5, got: %v", body["result"])
	}
}

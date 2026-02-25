package tests

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/sahina/cvt/sdks/go/cvt"
	"github.com/sahina/cvt/sdks/go/cvt/adapters"
)

// TestMock_AddResponse verifies mock generates valid response for /add
func TestMock_AddResponse(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	mock := adapters.NewMock(validator)
	client := mock.Client()

	resp, err := client.Get("http://calculator-api/add")
	if err != nil {
		t.Fatalf("Mock fetch failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if _, ok := data["result"]; !ok {
		t.Errorf("Expected 'result' field in response, got: %v", data)
	}
}

// TestMock_SubtractResponse verifies mock generates valid response for /subtract
func TestMock_SubtractResponse(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	mock := adapters.NewMock(validator)
	client := mock.Client()

	resp, err := client.Get("http://calculator-api/subtract")
	if err != nil {
		t.Fatalf("Mock fetch failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if _, ok := data["result"]; !ok {
		t.Errorf("Expected 'result' field in response, got: %v", data)
	}
}

// TestMock_CapturesInteractions verifies interactions are recorded
func TestMock_CapturesInteractions(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	mock := adapters.NewMock(validator)
	mock.ClearInteractions()
	client := mock.Client()

	client.Get("http://calculator-api/add")

	interactions := mock.GetInteractions()
	if len(interactions) != 1 {
		t.Errorf("Expected 1 interaction, got %d", len(interactions))
	}
}

// TestMock_CapturesAllConsumer4Endpoints verifies both add+subtract endpoints are recorded
func TestMock_CapturesAllConsumer4Endpoints(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	mock := adapters.NewMock(validator)
	mock.ClearInteractions()
	client := mock.Client()

	client.Get("http://calculator-api/add")
	client.Get("http://calculator-api/subtract")

	interactions := mock.GetInteractions()
	if len(interactions) != 2 {
		t.Errorf("Expected 2 interactions, got %d", len(interactions))
	}
}

// TestMock_ResponseValidatesAgainstSchema verifies mock response matches schema
func TestMock_ResponseValidatesAgainstSchema(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	mock := adapters.NewMock(validator)
	client := mock.Client()

	resp, err := client.Get("http://calculator-api/add")
	if err != nil {
		t.Fatalf("Mock fetch failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	var bodyData map[string]interface{}
	if err := json.Unmarshal(body, &bodyData); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	ctx := context.Background()
	result, err := validator.Validate(ctx,
		cvt.ValidationRequest{Method: "GET", Path: "/add"},
		cvt.ValidationResponse{StatusCode: 200, Body: bodyData},
	)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}
	if !result.Valid {
		t.Errorf("Mock response should validate against schema: %v", result.Errors)
	}
}

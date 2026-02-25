package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sahina/cvt/sdks/go/cvt/adapters"
)

// TestAdapter_AddAutoValidation tests auto-validation for add operation
func TestAdapter_AddAutoValidation(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	rt := adapters.NewValidatingRoundTripper(adapters.RoundTripperConfig{
		Validator:    validator,
		AutoValidate: true,
	})
	client := &http.Client{Transport: rt}

	url := fmt.Sprintf("%s/add?x=6&y=7", config.ProducerURL)
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	interactions := rt.GetInteractions()
	if len(interactions) == 0 {
		t.Error("Expected at least 1 interaction captured")
	}
	if interactions[0].ValidationResult == nil || !interactions[0].ValidationResult.Valid {
		errs := []string{}
		if interactions[0].ValidationResult != nil {
			errs = interactions[0].ValidationResult.Errors
		}
		t.Errorf("Expected valid interaction, got errors: %v", errs)
	}
}

// TestAdapter_SubtractAutoValidation tests auto-validation for subtract operation
func TestAdapter_SubtractAutoValidation(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	rt := adapters.NewValidatingRoundTripper(adapters.RoundTripperConfig{
		Validator:    validator,
		AutoValidate: true,
	})
	client := &http.Client{Transport: rt}
	defer rt.ClearInteractions()

	url := fmt.Sprintf("%s/subtract?x=10&y=4", config.ProducerURL)
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if val, ok := body["result"].(float64); !ok || val != 6 {
		t.Errorf("Expected result=6, got: %v", body["result"])
	}
}

// TestAdapter_CapturesMultipleInteractions verifies multiple interactions are recorded
func TestAdapter_CapturesMultipleInteractions(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	rt := adapters.NewValidatingRoundTripper(adapters.RoundTripperConfig{
		Validator:    validator,
		AutoValidate: true,
	})
	client := &http.Client{Transport: rt}

	client.Get(fmt.Sprintf("%s/add?x=3&y=4", config.ProducerURL))
	client.Get(fmt.Sprintf("%s/subtract?x=8&y=2", config.ProducerURL))

	interactions := rt.GetInteractions()
	if len(interactions) != 2 {
		t.Errorf("Expected 2 interactions, got %d", len(interactions))
	}
}

// TestAdapter_AllInteractionsValid verifies all captured interactions are valid
func TestAdapter_AllInteractionsValid(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	rt := adapters.NewValidatingRoundTripper(adapters.RoundTripperConfig{
		Validator:    validator,
		AutoValidate: true,
	})
	client := &http.Client{Transport: rt}

	client.Get(fmt.Sprintf("%s/add?x=5&y=5", config.ProducerURL))
	client.Get(fmt.Sprintf("%s/subtract?x=9&y=3", config.ProducerURL))

	interactions := rt.GetInteractions()
	for i, interaction := range interactions {
		if interaction.ValidationResult == nil || !interaction.ValidationResult.Valid {
			errs := []string{}
			if interaction.ValidationResult != nil {
				errs = interaction.ValidationResult.Errors
			}
			t.Errorf("Interaction %d failed validation: %v", i, errs)
		}
	}
}

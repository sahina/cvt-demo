package tests

import (
	"context"
	"testing"

	"github.com/sahina/cvt/sdks/go/cvt"
	"github.com/sahina/cvt/sdks/go/cvt/adapters"
)

// TestRegistration_CaptureInteractionsForAutoRegistration verifies auto-registration options can be built
func TestRegistration_CaptureInteractionsForAutoRegistration(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	mock := adapters.NewMock(validator)
	mock.ClearInteractions()
	client := mock.Client()

	client.Get("http://calculator-api/add")
	client.Get("http://calculator-api/subtract")

	interactions := mock.GetInteractions()
	ctx := context.Background()

	opts, err := validator.BuildConsumerFromInteractions(ctx, interactions, cvt.AutoRegisterConfig{
		ConsumerID:      config.ConsumerID,
		ConsumerVersion: config.Version,
		Environment:     config.Environment,
		SchemaVersion:   "1.0.0",
		SchemaID:        "calculator-api",
	})
	if err != nil {
		t.Fatalf("BuildConsumerFromInteractions failed: %v", err)
	}

	if opts == nil {
		t.Fatal("Expected non-nil options")
	}
	if opts.ConsumerID != config.ConsumerID {
		t.Errorf("Expected ConsumerID=%s, got %s", config.ConsumerID, opts.ConsumerID)
	}
	if len(opts.UsedEndpoints) != 2 {
		t.Errorf("Expected 2 used endpoints, got %d", len(opts.UsedEndpoints))
	}
}

// TestRegistration_RegisterConsumerFromInteractions registers consumer from mock interactions
func TestRegistration_RegisterConsumerFromInteractions(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	mock := adapters.NewMock(validator)
	mock.ClearInteractions()
	client := mock.Client()

	client.Get("http://calculator-api/add")
	client.Get("http://calculator-api/subtract")

	interactions := mock.GetInteractions()
	ctx := context.Background()

	consumer, err := validator.RegisterConsumerFromInteractions(ctx, interactions, cvt.AutoRegisterConfig{
		ConsumerID:      config.ConsumerID,
		ConsumerVersion: config.Version,
		Environment:     config.Environment,
		SchemaVersion:   "1.0.0",
		SchemaID:        "calculator-api",
	})
	if err != nil {
		t.Fatalf("RegisterConsumerFromInteractions failed: %v", err)
	}

	if consumer == nil {
		t.Fatal("Expected non-nil consumer info")
	}
	if consumer.ConsumerID != config.ConsumerID {
		t.Errorf("Expected ConsumerID=%s, got %s", config.ConsumerID, consumer.ConsumerID)
	}
}

// TestRegistration_RegisterConsumerWithExplicitEndpoints registers consumer with explicit endpoint list
func TestRegistration_RegisterConsumerWithExplicitEndpoints(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()

	consumer, err := validator.RegisterConsumer(ctx, cvt.RegisterConsumerOptions{
		ConsumerID:      config.ConsumerID,
		ConsumerVersion: config.Version,
		SchemaID:        "calculator-api",
		SchemaVersion:   "1.0.0",
		Environment:     config.Environment,
		UsedEndpoints: []cvt.EndpointUsage{
			{Method: "GET", Path: "/add", UsedFields: []string{"result"}},
			{Method: "GET", Path: "/subtract", UsedFields: []string{"result"}},
		},
	})
	if err != nil {
		t.Fatalf("RegisterConsumer failed: %v", err)
	}

	if consumer == nil {
		t.Fatal("Expected non-nil consumer info")
	}
	if consumer.ConsumerID != config.ConsumerID {
		t.Errorf("Expected ConsumerID=%s, got %s", config.ConsumerID, consumer.ConsumerID)
	}
}

// TestRegistration_ListRegisteredConsumers verifies registered consumers can be listed
func TestRegistration_ListRegisteredConsumers(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()

	// Register first to ensure at least one consumer exists
	_, err := validator.RegisterConsumer(ctx, cvt.RegisterConsumerOptions{
		ConsumerID:      config.ConsumerID,
		ConsumerVersion: config.Version,
		SchemaID:        "calculator-api",
		SchemaVersion:   "1.0.0",
		Environment:     config.Environment,
		UsedEndpoints: []cvt.EndpointUsage{
			{Method: "GET", Path: "/add", UsedFields: []string{"result"}},
			{Method: "GET", Path: "/subtract", UsedFields: []string{"result"}},
		},
	})
	if err != nil {
		t.Fatalf("RegisterConsumer failed: %v", err)
	}

	consumers, err := validator.ListConsumers(ctx, "calculator-api", config.Environment)
	if err != nil {
		t.Fatalf("ListConsumers failed: %v", err)
	}

	if len(consumers) == 0 {
		t.Error("Expected at least one registered consumer")
	}

	found := false
	for _, c := range consumers {
		if c.ConsumerID == config.ConsumerID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find consumer %s in list", config.ConsumerID)
	}
}

// TestRegistration_CanIDeploy checks deployment safety for current schema version
func TestRegistration_CanIDeploy(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()

	// Register consumer to ensure data exists
	_, err := validator.RegisterConsumer(ctx, cvt.RegisterConsumerOptions{
		ConsumerID:      config.ConsumerID,
		ConsumerVersion: config.Version,
		SchemaID:        "calculator-api",
		SchemaVersion:   "1.0.0",
		Environment:     config.Environment,
		UsedEndpoints: []cvt.EndpointUsage{
			{Method: "GET", Path: "/add", UsedFields: []string{"result"}},
			{Method: "GET", Path: "/subtract", UsedFields: []string{"result"}},
		},
	})
	if err != nil {
		t.Fatalf("RegisterConsumer failed: %v", err)
	}

	result, err := validator.CanIDeploy(ctx, "calculator-api", "1.0.0", config.Environment)
	if err != nil {
		t.Fatalf("CanIDeploy failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil CanIDeployResult")
	}
	// SafeToDeploy should be a valid boolean (not checking the value itself
	// as it depends on server state)
	t.Logf("CanIDeploy result: SafeToDeploy=%v, Summary=%s", result.SafeToDeploy, result.Summary)
}

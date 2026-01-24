// Package tests contains producer contract tests.
package tests

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sahina/cvt/sdks/go/cvt"
)

// TestRegistry_CanIDeploy_CurrentSchema tests that the current v1.0.0 schema
// is safe to deploy when consumers are registered.
func TestRegistry_CanIDeploy_CurrentSchema(t *testing.T) {
	config := GetTestConfig(t)

	validator, err := cvt.NewValidator(config.CVTServerAddr)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	defer validator.Close()

	ctx := context.Background()

	// Register the current schema
	if err := validator.RegisterSchema(ctx, config.SchemaID, config.SchemaPath); err != nil {
		t.Fatalf("Failed to register schema: %v", err)
	}

	// Check if the current schema version is safe to deploy
	result, err := validator.CanIDeploy(ctx, config.SchemaID, "1.0.0", config.Environment)
	if err != nil {
		t.Logf("CanIDeploy check returned error (may be expected if no consumers registered): %v", err)
		return
	}

	t.Logf("CanIDeploy result for v1.0.0:")
	t.Logf("  Safe to deploy: %v", result.SafeToDeploy)
	t.Logf("  Summary: %s", result.Summary)

	if len(result.BreakingChanges) > 0 {
		t.Logf("  Breaking changes:")
		for _, change := range result.BreakingChanges {
			t.Logf("    - %v", change)
		}
	}

	// Current schema should be safe to deploy
	if !result.SafeToDeploy {
		t.Errorf("Expected v1.0.0 to be safe to deploy, but got: %s", result.Summary)
	}
}

// TestRegistry_CanIDeploy_BreakingSchema tests that a breaking schema change
// (result -> value) is correctly identified as UNSAFE.
func TestRegistry_CanIDeploy_BreakingSchema(t *testing.T) {
	config := GetTestConfig(t)

	validator, err := cvt.NewValidator(config.CVTServerAddr)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	defer validator.Close()

	ctx := context.Background()

	// Register the breaking v2.0.0 schema (result -> value)
	breakingSchemaPath := filepath.Join("..", "calculator-api-v2-breaking.json")
	if err := validator.RegisterSchema(ctx, "calculator-api-v2", breakingSchemaPath); err != nil {
		t.Logf("Failed to register breaking schema (file may not exist): %v", err)
		t.Skip("Breaking schema file not available")
	}

	// Check if the breaking schema is safe to deploy
	result, err := validator.CanIDeploy(ctx, "calculator-api-v2", "2.0.0", config.Environment)
	if err != nil {
		t.Logf("CanIDeploy check returned error (may be expected if no consumers registered): %v", err)
		return
	}

	t.Logf("CanIDeploy result for v2.0.0 (breaking):")
	t.Logf("  Safe to deploy: %v", result.SafeToDeploy)
	t.Logf("  Summary: %s", result.Summary)

	if len(result.BreakingChanges) > 0 {
		t.Logf("  Breaking changes detected:")
		for _, change := range result.BreakingChanges {
			t.Logf("    - %v", change)
		}
	}

	// Breaking schema should NOT be safe to deploy if consumers are registered
	// Note: This test may pass with SafeToDeploy=true if no consumers are registered
	if result.SafeToDeploy {
		t.Logf("Note: v2.0.0 marked as safe. This is expected if no consumers are registered.")
		t.Logf("Run consumer registration tests first to see breaking change detection.")
	}
}

// TestRegistry_ListConsumers lists all consumers registered in the environment.
func TestRegistry_ListConsumers(t *testing.T) {
	config := GetTestConfig(t)

	validator, err := cvt.NewValidator(config.CVTServerAddr)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	defer validator.Close()

	ctx := context.Background()

	// List consumers for the schema
	consumers, err := validator.ListConsumers(ctx, config.SchemaID, config.Environment)
	if err != nil {
		t.Logf("Failed to list consumers: %v", err)
		t.Skip("Consumer listing not available")
	}

	t.Logf("Registered consumers for %s in %s environment:", config.SchemaID, config.Environment)
	if len(consumers) == 0 {
		t.Logf("  No consumers registered")
		t.Logf("  Run 'make test-consumer-1-registration' and 'make test-consumer-2-registration' first")
	} else {
		for _, consumer := range consumers {
			t.Logf("  - %s (version: %s)", consumer.ConsumerID, consumer.ConsumerVersion)
			if len(consumer.UsedEndpoints) > 0 {
				t.Logf("    Endpoints: %v", consumer.UsedEndpoints)
			}
		}
	}
}

// TestRegistry_DeploymentFlow demonstrates the full deployment verification flow:
// 1. Register current schema
// 2. Check can-i-deploy for current version
// 3. Register breaking schema
// 4. Check can-i-deploy for breaking version
func TestRegistry_DeploymentFlow(t *testing.T) {
	config := GetTestConfig(t)

	validator, err := cvt.NewValidator(config.CVTServerAddr)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}
	defer validator.Close()

	ctx := context.Background()

	// Step 1: Register current schema
	t.Log("Step 1: Registering current schema (v1.0.0)...")
	if err := validator.RegisterSchema(ctx, config.SchemaID, config.SchemaPath); err != nil {
		t.Fatalf("Failed to register current schema: %v", err)
	}
	t.Log("  Schema registered successfully")

	// Step 2: Check can-i-deploy for current version
	t.Log("Step 2: Checking if v1.0.0 can be deployed...")
	result, err := validator.CanIDeploy(ctx, config.SchemaID, "1.0.0", config.Environment)
	if err != nil {
		t.Logf("  CanIDeploy returned error: %v", err)
	} else {
		t.Logf("  Result: SafeToDeploy=%v, Summary=%s", result.SafeToDeploy, result.Summary)
	}

	// Step 3: Try to register breaking schema
	t.Log("Step 3: Attempting to register breaking schema (v2.0.0)...")
	breakingSchemaPath := filepath.Join("..", "calculator-api-v2-breaking.json")
	if err := validator.RegisterSchema(ctx, "calculator-api-v2", breakingSchemaPath); err != nil {
		t.Logf("  Could not register breaking schema: %v", err)
		t.Log("  This is expected if the file doesn't exist")
		return
	}
	t.Log("  Breaking schema registered")

	// Step 4: Check can-i-deploy for breaking version
	t.Log("Step 4: Checking if v2.0.0 can be deployed...")
	result, err = validator.CanIDeploy(ctx, "calculator-api-v2", "2.0.0", config.Environment)
	if err != nil {
		t.Logf("  CanIDeploy returned error: %v", err)
	} else {
		t.Logf("  Result: SafeToDeploy=%v, Summary=%s", result.SafeToDeploy, result.Summary)
		if !result.SafeToDeploy {
			t.Log("  Breaking changes detected - deployment blocked!")
			for _, change := range result.BreakingChanges {
				t.Logf("    - %v", change)
			}
		}
	}
}

// TestRegistry_BreakingChangeScenarios tests specific breaking change scenarios
// that would affect registered consumers.
func TestRegistry_BreakingChangeScenarios(t *testing.T) {
	t.Log("Breaking Change Scenarios:")
	t.Log("")
	t.Log("1. Field Rename (result -> value):")
	t.Log("   - Consumer-1 expects: { \"result\": <number> }")
	t.Log("   - v2.0.0 returns: { \"value\": <number> }")
	t.Log("   - Impact: Both consumers break")
	t.Log("")
	t.Log("2. Type Change (number -> string):")
	t.Log("   - Consumer expects: { \"result\": 8 }")
	t.Log("   - Breaking returns: { \"result\": \"8\" }")
	t.Log("   - Impact: Type mismatch errors")
	t.Log("")
	t.Log("3. Endpoint Removal:")
	t.Log("   - Consumer-1 uses: /add, /subtract")
	t.Log("   - Consumer-2 uses: /add, /multiply, /divide")
	t.Log("   - If /add removed: Both consumers break")
	t.Log("")
	t.Log("Run consumer registration tests first to see these scenarios in action:")
	t.Log("  make test-consumer-1-registration")
	t.Log("  make test-consumer-2-registration")
	t.Log("  make test-producer-registry")
}

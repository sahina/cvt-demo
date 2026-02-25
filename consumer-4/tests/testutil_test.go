package tests

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sahina/cvt/sdks/go/cvt"
)

type testConfig struct {
	CVTServerAddr string
	ProducerURL   string
	SchemaPath    string
	ConsumerID    string
	Version       string
	Environment   string
}

func getTestConfig(t *testing.T) *testConfig {
	t.Helper()

	cvtAddr := os.Getenv("CVT_SERVER_ADDR")
	if cvtAddr == "" {
		cvtAddr = "localhost:9550"
	}

	producerURL := os.Getenv("PRODUCER_URL")
	if producerURL == "" {
		producerURL = "http://localhost:10001"
	}

	schemaPath := os.Getenv("SCHEMA_PATH")
	if schemaPath == "" {
		_, filename, _, _ := runtime.Caller(0)
		schemaPath = filepath.Join(filepath.Dir(filename), "../../producer/calculator-api.json")
	}

	env := os.Getenv("CVT_ENVIRONMENT")
	if env == "" {
		env = "demo"
	}

	return &testConfig{
		CVTServerAddr: cvtAddr,
		ProducerURL:   producerURL,
		SchemaPath:    schemaPath,
		ConsumerID:    "consumer-4",
		Version:       "1.0.0",
		Environment:   env,
	}
}

func newTestValidator(t *testing.T, config *testConfig) *cvt.Validator {
	t.Helper()

	validator, err := cvt.NewValidator(config.CVTServerAddr)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	ctx := context.Background()
	if err := validator.RegisterSchema(ctx, "calculator-api", config.SchemaPath); err != nil {
		validator.Close()
		t.Fatalf("Failed to register schema: %v", err)
	}

	return validator
}

func headersToMap(h http.Header) map[string]string {
	result := make(map[string]string)
	for k, v := range h {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

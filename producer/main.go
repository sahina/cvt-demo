// Package main implements a calculator API with CVT middleware for contract validation.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sahina/cvt-demo/producer/handlers"
	"github.com/sahina/cvt/sdks/go/cvt"
	"github.com/sahina/cvt/sdks/go/cvt/producer"
	"github.com/sahina/cvt/sdks/go/cvt/producer/adapters"
)

// validatorAdapter adapts cvt.Validator to producer.Validator interface.
type validatorAdapter struct {
	validator *cvt.Validator
}

// Validate implements the producer.Validator interface.
func (a *validatorAdapter) Validate(ctx context.Context, schemaID string, interaction *producer.Interaction) (*producer.ValidationResult, error) {
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
	result, err := a.validator.Validate(ctx, request, response)
	if err != nil {
		return nil, err
	}

	return &producer.ValidationResult{
		Valid:  result.Valid,
		Errors: result.Errors,
	}, nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10001"
	}

	cvtServerAddr := os.Getenv("CVT_SERVER_ADDR")
	if cvtServerAddr == "" {
		cvtServerAddr = "localhost:9550"
	}

	schemaPath := os.Getenv("SCHEMA_PATH")
	if schemaPath == "" {
		schemaPath = "./calculator-api.yaml"
	}

	// Create the HTTP mux
	mux := http.NewServeMux()

	// Register handlers
	calc := handlers.NewCalculator()
	calc.RegisterRoutes(mux)

	// Determine if CVT validation is enabled
	cvtEnabled := os.Getenv("CVT_ENABLED") != "false"

	var handler http.Handler = mux

	if cvtEnabled {
		// Create CVT validator
		validator, err := cvt.NewValidator(cvtServerAddr)
		if err != nil {
			log.Printf("Warning: Failed to create CVT validator: %v. Running without validation.", err)
		} else {
			// Register schema
			ctx := context.Background()
			if err := validator.RegisterSchema(ctx, "calculator-api", schemaPath); err != nil {
				log.Printf("Warning: Failed to register schema: %v. Running without validation.", err)
			} else {
				log.Printf("CVT validation enabled with schema from %s", schemaPath)

				// Create adapter that implements producer.Validator
				adapter := &validatorAdapter{validator: validator}

				// Create producer config
				config := producer.Config{
					SchemaID:         "calculator-api",
					Validator:        adapter,
					Mode:             producer.ModeStrict,
					ValidateRequest:  true,
					ValidateResponse: true,
					ExcludePaths:     []producer.PathFilter{"/health"},
				}

				// Wrap with CVT middleware
				handler = adapters.NetHTTPMiddleware(config)(mux)
			}
		}
	} else {
		log.Println("CVT validation disabled")
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Calculator API starting on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

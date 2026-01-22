// Package main implements a calculator API with CVT middleware for contract validation.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/sahina/cvt/sdks/go/cvt"
	"github.com/sahina/cvt/sdks/go/cvt/producer"
	"github.com/sahina/cvt/sdks/go/cvt/producer/adapters"
)

// ResultResponse represents a successful calculation result.
type ResultResponse struct {
	Result float64 `json:"result"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

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
	mux.HandleFunc("/add", handleAdd)
	mux.HandleFunc("/subtract", handleSubtract)
	mux.HandleFunc("/multiply", handleMultiply)
	mux.HandleFunc("/divide", handleDivide)
	mux.HandleFunc("/health", handleHealth)

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

// parseNumbers extracts and validates query parameters a and b.
func parseNumbers(w http.ResponseWriter, r *http.Request) (float64, float64, bool) {
	aStr := r.URL.Query().Get("a")
	bStr := r.URL.Query().Get("b")

	if aStr == "" || bStr == "" {
		writeError(w, "missing required parameters 'a' and 'b'", http.StatusBadRequest)
		return 0, 0, false
	}

	a, err := strconv.ParseFloat(aStr, 64)
	if err != nil {
		writeError(w, "parameter 'a' must be a valid number", http.StatusBadRequest)
		return 0, 0, false
	}

	b, err := strconv.ParseFloat(bStr, 64)
	if err != nil {
		writeError(w, "parameter 'b' must be a valid number", http.StatusBadRequest)
		return 0, 0, false
	}

	return a, b, true
}

// writeResult writes a successful result response.
func writeResult(w http.ResponseWriter, result float64) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResultResponse{Result: result})
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// handleAdd handles the /add endpoint.
func handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a, b, ok := parseNumbers(w, r)
	if !ok {
		return
	}

	writeResult(w, a+b)
}

// handleSubtract handles the /subtract endpoint.
func handleSubtract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a, b, ok := parseNumbers(w, r)
	if !ok {
		return
	}

	writeResult(w, a-b)
}

// handleMultiply handles the /multiply endpoint.
func handleMultiply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a, b, ok := parseNumbers(w, r)
	if !ok {
		return
	}

	writeResult(w, a*b)
}

// handleDivide handles the /divide endpoint.
func handleDivide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	a, b, ok := parseNumbers(w, r)
	if !ok {
		return
	}

	if b == 0 {
		writeError(w, "division by zero is not allowed", http.StatusBadRequest)
		return
	}

	writeResult(w, a/b)
}

// handleHealth handles the /health endpoint.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// Package handlers provides HTTP handlers for the Calculator API.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// ResultResponse represents a successful calculation result.
type ResultResponse struct {
	Result float64 `json:"result"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// HealthResponse represents a health check response.
type HealthResponse struct {
	Status string `json:"status"`
}

// Calculator handles all calculator operations.
type Calculator struct{}

// NewCalculator creates a new Calculator instance.
func NewCalculator() *Calculator {
	return &Calculator{}
}

// RegisterRoutes registers all calculator routes on the given mux.
func (c *Calculator) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/add", c.Add)
	mux.HandleFunc("/subtract", c.Subtract)
	mux.HandleFunc("/multiply", c.Multiply)
	mux.HandleFunc("/divide", c.Divide)
	mux.HandleFunc("/health", c.Health)
}

// Add handles the /add endpoint.
func (c *Calculator) Add(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	x, y, ok := parseNumbers(w, r)
	if !ok {
		return
	}

	writeResult(w, x+y)
}

// Subtract handles the /subtract endpoint.
func (c *Calculator) Subtract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	x, y, ok := parseNumbers(w, r)
	if !ok {
		return
	}

	writeResult(w, x-y)
}

// Multiply handles the /multiply endpoint.
func (c *Calculator) Multiply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	x, y, ok := parseNumbers(w, r)
	if !ok {
		return
	}

	writeResult(w, x*y)
}

// Divide handles the /divide endpoint.
func (c *Calculator) Divide(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	x, y, ok := parseNumbers(w, r)
	if !ok {
		return
	}

	if y == 0 {
		writeError(w, "division by zero is not allowed", http.StatusBadRequest)
		return
	}

	writeResult(w, x/y)
}

// Health handles the /health endpoint.
func (c *Calculator) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{Status: "healthy"})
}

// parseNumbers extracts and validates query parameters x and y.
func parseNumbers(w http.ResponseWriter, r *http.Request) (float64, float64, bool) {
	xStr := r.URL.Query().Get("x")
	yStr := r.URL.Query().Get("y")

	if xStr == "" || yStr == "" {
		writeError(w, "missing required parameters 'x' and 'y'", http.StatusBadRequest)
		return 0, 0, false
	}

	x, err := strconv.ParseFloat(xStr, 64)
	if err != nil {
		writeError(w, "parameter 'x' must be a valid number", http.StatusBadRequest)
		return 0, 0, false
	}

	y, err := strconv.ParseFloat(yStr, 64)
	if err != nil {
		writeError(w, "parameter 'y' must be a valid number", http.StatusBadRequest)
		return 0, 0, false
	}

	return x, y, true
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

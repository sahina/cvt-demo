// consumer-4: A CLI tool that uses the Calculator API for add and subtract operations.
//
// Usage:
//
//	./consumer4 add <x> <y> [--validate]
//	./consumer4 subtract <x> <y> [--validate]
//
// Options:
//
//	--validate  Enable CVT contract validation (default: off)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/sahina/cvt/sdks/go/cvt"
)

var (
	producerURL = getEnv("PRODUCER_URL", "http://localhost:10001")
	cvtAddr     = getEnv("CVT_SERVER_ADDR", "localhost:9550")
	schemaPath  = getEnv("SCHEMA_PATH", "./calculator-api.json")
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "Usage: consumer4 <add|subtract> <x> <y> [--validate]")
		os.Exit(1)
	}

	command := os.Args[1]
	if command != "add" && command != "subtract" {
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'. Use 'add' or 'subtract'.\n", command)
		os.Exit(1)
	}

	x, err := strconv.ParseFloat(os.Args[2], 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Both arguments must be valid numbers")
		os.Exit(1)
	}
	y, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Both arguments must be valid numbers")
		os.Exit(1)
	}

	validate := len(os.Args) > 4 && os.Args[4] == "--validate"

	path := fmt.Sprintf("/%s?x=%s&y=%s", command, formatParam(x), formatParam(y))
	url := producerURL + path

	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		if json.Unmarshal(bodyBytes, &errResp) == nil {
			if msg, ok := errResp["error"].(string); ok {
				fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
				os.Exit(1)
			}
		}
		fmt.Fprintf(os.Stderr, "Error: HTTP %d\n", resp.StatusCode)
		os.Exit(1)
	}

	if validate {
		runValidation(path, resp.StatusCode, bodyBytes)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing response: %v\n", err)
		os.Exit(1)
	}

	val, ok := result["result"].(float64)
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: Unexpected response format: %s\n", string(bodyBytes))
		os.Exit(1)
	}

	op := "+"
	if command == "subtract" {
		op = "-"
	}
	fmt.Printf("%s %s %s = %s\n", formatParam(x), op, formatParam(y), formatNumber(val))
}

// runValidation validates the interaction with CVT.
// Prints a warning and continues execution if CVT is unreachable (graceful fallback).
func runValidation(path string, statusCode int, body []byte) {
	if err := doValidation(path, statusCode, body); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to enable CVT validation: %v\n", err)
		fmt.Fprintln(os.Stderr, "Continuing without validation...")
	}
}

func doValidation(path string, statusCode int, body []byte) error {
	validator, err := cvt.NewValidator(cvtAddr)
	if err != nil {
		return err
	}
	defer validator.Close()

	ctx := context.Background()
	if err := validator.RegisterSchema(ctx, "calculator-api", schemaPath); err != nil {
		return err
	}

	var bodyData map[string]interface{}
	json.Unmarshal(body, &bodyData) //nolint:errcheck

	result, err := validator.Validate(ctx,
		cvt.ValidationRequest{Method: "GET", Path: path},
		cvt.ValidationResponse{StatusCode: statusCode, Body: bodyData},
	)
	if err != nil {
		return err
	}

	if !result.Valid {
		fmt.Fprintf(os.Stderr, "CVT Validation failed: %v\n", result.Errors)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "CVT validation passed")
	return nil
}

func formatParam(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func formatNumber(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

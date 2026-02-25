package demo.consumer3;

import io.github.sahina.sdk.ContractValidator;
import io.github.sahina.sdk.ValidationRequest;
import io.github.sahina.sdk.ValidationResponse;
import io.github.sahina.sdk.ValidationResult;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.List;
import java.util.Map;

/**
 * Consumer-3: A CLI tool that uses the Calculator API for multiply and divide operations.
 *
 * Usage:
 *   java -jar consumer3.jar multiply <x> <y> [--validate]
 *   java -jar consumer3.jar divide <x> <y> [--validate]
 *
 * Options:
 *   --validate  Enable CVT contract validation (default: off)
 */
public class Main {

    private static final String PRODUCER_URL =
            System.getenv().getOrDefault("PRODUCER_URL", "http://localhost:10001");
    private static final String CVT_SERVER_ADDR =
            System.getenv().getOrDefault("CVT_SERVER_ADDR", "localhost:9550");
    private static final String SCHEMA_PATH =
            System.getenv().getOrDefault("SCHEMA_PATH", "./calculator-api.json");

    public static void main(String[] args) {
        if (args.length < 3) {
            System.err.println("Usage: java -jar consumer3.jar <multiply|divide> <x> <y> [--validate]");
            System.exit(1);
        }

        String command = args[0];
        if (!command.equals("multiply") && !command.equals("divide")) {
            System.err.println("Error: Unknown command '" + command + "'. Use 'multiply' or 'divide'.");
            System.exit(1);
        }

        double x;
        double y;
        try {
            x = Double.parseDouble(args[1]);
            y = Double.parseDouble(args[2]);
        } catch (NumberFormatException e) {
            System.err.println("Error: Both arguments must be valid numbers");
            System.exit(1);
            return; // satisfy compiler
        }

        boolean validate = args.length >= 4 && "--validate".equals(args[3]);

        try {
            run(command, x, y, validate);
        } catch (Exception e) {
            System.err.println("Error: " + e.getMessage());
            System.exit(1);
        }
    }

    private static void run(String command, double x, double y, boolean validate) throws Exception {
        String path = "/" + command + "?x=" + formatParam(x) + "&y=" + formatParam(y);
        String url = PRODUCER_URL + path;

        HttpClient client = HttpClient.newBuilder()
                .connectTimeout(java.time.Duration.ofSeconds(10))
                .build();
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(url))
                .header("Accept", "application/json")
                .GET()
                .build();

        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());

        String body = response.body();
        int statusCode = response.statusCode();

        if (statusCode >= 400) {
            String errorMsg = extractJsonField(body, "error");
            if (errorMsg != null) {
                System.err.println("Error: " + errorMsg);
            } else {
                System.err.println("Error: HTTP " + statusCode);
            }
            System.exit(1);
        }

        if (validate) {
            try {
                try (ContractValidator validator = new ContractValidator(CVT_SERVER_ADDR)) {
                    validator.registerSchema("calculator-api", SCHEMA_PATH);

                    ValidationRequest validationRequest = ValidationRequest.builder()
                            .method("GET")
                            .path(path)
                            .headers(Map.of())
                            .build();

                    ValidationResponse validationResponse = ValidationResponse.builder()
                            .statusCode(statusCode)
                            .header("content-type", "application/json")
                            .body(body)
                            .build();

                    ValidationResult result = validator.validate(validationRequest, validationResponse);

                    if (!result.isValid()) {
                        List<String> errors = result.getErrors();
                        System.err.println("CVT Validation failed: " + String.join(", ", errors));
                        System.exit(1);
                    }
                }
            } catch (Exception e) {
                System.err.println("Warning: Failed to enable CVT validation: " + e.getMessage());
                System.err.println("Continuing without validation...");
            }
        }

        double resultValue = parseResultFromBody(body);
        String formatted = formatResult(x, command, y, resultValue);
        System.out.println(formatted);
    }

    /**
     * Formats a number for use as a query parameter.
     * If the value is a whole number, omit the decimal point.
     */
    private static String formatParam(double value) {
        if (value == Math.floor(value) && !Double.isInfinite(value)) {
            return String.valueOf((long) value);
        }
        return String.valueOf(value);
    }

    /**
     * Parses the "result" field from a JSON body like {"result":28} or {"result":5.0}.
     * Uses simple string manipulation — no external JSON library.
     */
    private static double parseResultFromBody(String body) {
        String numStr = extractJsonField(body, "result");
        if (numStr == null) {
            throw new RuntimeException("Could not parse 'result' from response: " + body);
        }
        return Double.parseDouble(numStr);
    }

    /**
     * Extracts a JSON field value by key from a simple flat JSON object.
     * Returns the raw string value (without quotes for strings, as-is for numbers).
     */
    private static String extractJsonField(String body, String key) {
        if (body == null || body.isEmpty()) {
            return null;
        }
        String searchKey = "\"" + key + "\"";
        int keyIdx = body.indexOf(searchKey);
        if (keyIdx < 0) {
            return null;
        }
        int colonIdx = body.indexOf(':', keyIdx + searchKey.length());
        if (colonIdx < 0) {
            return null;
        }
        // Skip whitespace after colon
        int valueStart = colonIdx + 1;
        while (valueStart < body.length() && Character.isWhitespace(body.charAt(valueStart))) {
            valueStart++;
        }
        if (valueStart >= body.length()) {
            return null;
        }
        char firstChar = body.charAt(valueStart);
        if (firstChar == '"') {
            // String value — find closing quote
            int valueEnd = body.indexOf('"', valueStart + 1);
            if (valueEnd < 0) {
                return null;
            }
            return body.substring(valueStart + 1, valueEnd);
        } else {
            // Numeric or boolean value — read until delimiter
            int valueEnd = valueStart;
            while (valueEnd < body.length()) {
                char c = body.charAt(valueEnd);
                if (c == ',' || c == '}' || c == ']' || Character.isWhitespace(c)) {
                    break;
                }
                valueEnd++;
            }
            return body.substring(valueStart, valueEnd);
        }
    }

    /**
     * Formats the output line. Numbers are displayed as integers when they are whole numbers.
     * Examples: "4 * 7 = 28", "10 / 2 = 5", "10 / 3 = 3.3333333333333335"
     */
    private static String formatResult(double x, String command, double y, double result) {
        String xStr = formatNumber(x);
        String yStr = formatNumber(y);
        String resultStr = formatNumber(result);
        String operator = command.equals("multiply") ? "*" : "/";
        return xStr + " " + operator + " " + yStr + " = " + resultStr;
    }

    /**
     * Formats a double as an integer string if it is a whole number, otherwise as a decimal.
     */
    private static String formatNumber(double value) {
        if (value == Math.floor(value) && !Double.isInfinite(value) && !Double.isNaN(value)) {
            return String.valueOf((long) value);
        }
        return String.valueOf(value);
    }
}

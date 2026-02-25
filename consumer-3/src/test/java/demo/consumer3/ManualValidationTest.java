package demo.consumer3;

import io.github.sahina.sdk.ContractValidator;
import io.github.sahina.sdk.ValidationRequest;
import io.github.sahina.sdk.ValidationResponse;
import io.github.sahina.sdk.ValidationResult;

import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;

import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;

import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;

/**
 * MANUAL VALIDATION APPROACH
 * --------------------------
 * Makes real HTTP calls to the producer, then explicitly calls validator.validate().
 *
 * Prerequisites:
 * - CVT server running
 * - Producer service running
 *
 * When to use:
 * - When you need full control over the validation lifecycle
 * - Integration testing against a live producer
 */
public class ManualValidationTest {

    private static final String CVT_SERVER_ADDR =
            System.getenv().getOrDefault("CVT_SERVER_ADDR", "localhost:9550");
    private static final String PRODUCER_URL =
            System.getenv().getOrDefault("PRODUCER_URL", "http://localhost:10001");
    private static final String SCHEMA_PATH =
            System.getenv().getOrDefault("SCHEMA_PATH", "../producer/calculator-api.json");
    private static final String SCHEMA_ID = "calculator-api";

    private static ContractValidator validator;
    private static OkHttpClient httpClient;

    @BeforeAll
    static void setUp() throws Exception {
        validator = new ContractValidator(CVT_SERVER_ADDR);
        validator.registerSchema(SCHEMA_ID, SCHEMA_PATH);
        httpClient = new OkHttpClient.Builder()
                .connectTimeout(java.time.Duration.ofSeconds(10))
                .readTimeout(java.time.Duration.ofSeconds(10))
                .build();
    }

    @AfterAll
    static void tearDown() throws Exception {
        if (validator != null) {
            validator.close();
        }
    }

    @Test
    void shouldValidateSuccessfulMultiplyOperation() throws Exception {
        Request request = new Request.Builder()
                .url(PRODUCER_URL + "/multiply?x=4&y=7")
                .build();

        try (Response response = httpClient.newCall(request).execute()) {
            assertEquals(200, response.code(), "Expected HTTP 200 for multiply");
            assertNotNull(response.body(), "Response body should not be null");
            String body = response.body().string();

            ValidationRequest validationRequest = ValidationRequest.builder()
                    .method("GET")
                    .path("/multiply?x=4&y=7")
                    .headers(Map.of())
                    .build();

            ValidationResponse validationResponse = ValidationResponse.builder()
                    .statusCode(200)
                    .header("content-type", "application/json")
                    .body(body)
                    .build();

            ValidationResult result = validator.validate(validationRequest, validationResponse);

            assertTrue(result.isValid(),
                    "Response should be valid against schema. Errors: " + result.getErrors());
            assertTrue(body.contains("\"result\":28") || body.contains("\"result\":28.0") || body.contains("\"result\": 28"),
                    "Response body should contain result=28, got: " + body);
        }
    }

    @Test
    void shouldValidateSuccessfulDivideOperation() throws Exception {
        Request request = new Request.Builder()
                .url(PRODUCER_URL + "/divide?x=10&y=2")
                .build();

        try (Response response = httpClient.newCall(request).execute()) {
            assertEquals(200, response.code(), "Expected HTTP 200 for divide");
            assertNotNull(response.body(), "Response body should not be null");
            String body = response.body().string();

            ValidationRequest validationRequest = ValidationRequest.builder()
                    .method("GET")
                    .path("/divide?x=10&y=2")
                    .headers(Map.of())
                    .build();

            ValidationResponse validationResponse = ValidationResponse.builder()
                    .statusCode(200)
                    .header("content-type", "application/json")
                    .body(body)
                    .build();

            ValidationResult result = validator.validate(validationRequest, validationResponse);

            assertTrue(result.isValid(),
                    "Response should be valid against schema. Errors: " + result.getErrors());
            assertTrue(body.contains("\"result\":5") || body.contains("\"result\":5.0") || body.contains("\"result\": 5"),
                    "Response body should contain result=5, got: " + body);
        }
    }

    @Test
    void shouldDetectMissingResultField() throws Exception {
        // Synthetic response with wrong field name — no real HTTP call needed
        ValidationRequest validationRequest = ValidationRequest.builder()
                .method("GET")
                .path("/multiply?x=4&y=7")
                .headers(Map.of())
                .build();

        ValidationResponse validationResponse = ValidationResponse.builder()
                .statusCode(200)
                .header("content-type", "application/json")
                .body("{\"product\": 28}")
                .build();

        ValidationResult result = validator.validate(validationRequest, validationResponse);

        assertFalse(result.isValid(),
                "Response with wrong field name 'product' instead of 'result' should be invalid");
    }

    @Test
    void shouldValidateDivideByZeroErrorResponse() throws Exception {
        Request request = new Request.Builder()
                .url(PRODUCER_URL + "/divide?x=10&y=0")
                .build();

        try (Response response = httpClient.newCall(request).execute()) {
            assertEquals(400, response.code(), "Expected HTTP 400 for divide by zero");
            assertNotNull(response.body(), "Response body should not be null");
            String body = response.body().string();

            ValidationRequest validationRequest = ValidationRequest.builder()
                    .method("GET")
                    .path("/divide?x=10&y=0")
                    .headers(Map.of())
                    .build();

            ValidationResponse validationResponse = ValidationResponse.builder()
                    .statusCode(400)
                    .header("content-type", "application/json")
                    .body(body)
                    .build();

            ValidationResult result = validator.validate(validationRequest, validationResponse);

            assertTrue(result.isValid(),
                    "400 error response should match error schema. Errors: " + result.getErrors());
        }
    }
}

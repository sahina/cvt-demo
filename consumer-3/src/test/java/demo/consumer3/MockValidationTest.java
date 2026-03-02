package demo.consumer3;

import io.github.sahina.sdk.ContractValidator;
import io.github.sahina.sdk.ValidationRequest;
import io.github.sahina.sdk.ValidationResponse;
import io.github.sahina.sdk.ValidationResult;
import io.github.sahina.sdk.adapters.MockInterceptor;

import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;

import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;

/**
 * MOCK ADAPTER APPROACH
 * ---------------------
 * Generates mock responses based on the OpenAPI schema without making real HTTP calls.
 *
 * Prerequisites:
 * - CVT server running (for schema validation)
 * - No producer service needed
 *
 * When to use:
 * - Unit tests that need to run in isolation
 * - CI pipelines where the producer isn't available
 * - Testing consumer code paths without external dependencies
 */
public class MockValidationTest {

    private static final String CVT_SERVER_ADDR =
            System.getenv().getOrDefault("CVT_SERVER_ADDR", "localhost:9550");
    private static final String SCHEMA_PATH =
            System.getenv().getOrDefault("SCHEMA_PATH", "../producer/calculator-api.json");
    private static final String SCHEMA_ID = "calculator-api";

    private static ContractValidator validator;
    private static MockInterceptor mockInterceptor;
    private static OkHttpClient mockClient;

    @BeforeAll
    static void setUp() throws Exception {
        validator = new ContractValidator(CVT_SERVER_ADDR);
        validator.registerSchema(SCHEMA_ID, SCHEMA_PATH);

        mockInterceptor = new MockInterceptor(validator);
        mockClient = new OkHttpClient.Builder()
                .addInterceptor(mockInterceptor)
                .build();
    }

    @AfterAll
    static void tearDown() throws Exception {
        if (mockInterceptor != null) {
            mockInterceptor.clearCache();
        }
        if (validator != null) {
            validator.close();
        }
    }

    @BeforeEach
    void clearInteractions() {
        if (mockInterceptor != null) {
            mockInterceptor.clearInteractions();
        }
    }

    @Test
    void shouldGenerateValidMockResponseForMultiply() throws Exception {
        // Note: The Java MockInterceptor passes the full URL path (including query params)
        // to the CVT server for route matching, but the CVT server only knows routes without
        // query params (e.g., /multiply, not /multiply?x=4&y=7). We omit query params here
        // so the MockInterceptor sends "/multiply" to the CVT server, which can match it.
        Request request = new Request.Builder()
                .url("http://calculator-api/multiply")
                .build();

        try (Response response = mockClient.newCall(request).execute()) {
            assertEquals(200, response.code(), "Expected HTTP 200 for multiply mock response");
            assertNotNull(response.body(), "Response body should not be null");
            String body = response.body().string();
            assertTrue(body.contains("result"),
                    "Mock response body should contain 'result' field, got: " + body);
        }
    }

    @Test
    void shouldGenerateValidMockResponseForDivide() throws Exception {
        Request request = new Request.Builder()
                .url("http://calculator-api/divide")
                .build();

        try (Response response = mockClient.newCall(request).execute()) {
            assertEquals(200, response.code(), "Expected HTTP 200 for divide mock response");
            assertNotNull(response.body(), "Response body should not be null");
            String body = response.body().string();
            assertTrue(body.contains("result"),
                    "Mock response body should contain 'result' field, got: " + body);
        }
    }

    @Test
    void shouldCaptureMockInteractions() throws Exception {
        Request request = new Request.Builder()
                .url("http://calculator-api/multiply")
                .build();

        try (Response response = mockClient.newCall(request).execute()) {
            // consume body to complete the call
            assertNotNull(response.body());
            response.body().string();
        }

        List<?> interactions = mockInterceptor.getInteractions();
        assertEquals(1, interactions.size(),
                "Expected exactly 1 captured interaction after one mock request");
    }

    @Test
    void shouldCaptureAllConsumer3Endpoints() throws Exception {
        // Make multiply request
        Request multiplyRequest = new Request.Builder()
                .url("http://calculator-api/multiply")
                .build();
        try (Response response = mockClient.newCall(multiplyRequest).execute()) {
            assertNotNull(response.body());
            response.body().string();
        }

        // Make divide request
        Request divideRequest = new Request.Builder()
                .url("http://calculator-api/divide")
                .build();
        try (Response response = mockClient.newCall(divideRequest).execute()) {
            assertNotNull(response.body());
            response.body().string();
        }

        List<?> interactions = mockInterceptor.getInteractions();
        assertEquals(2, interactions.size(),
                "Expected 2 captured interactions for multiply and divide endpoints");
    }

    @Test
    void mockResponseShouldValidateAgainstSchema() throws Exception {
        Request request = new Request.Builder()
                .url("http://calculator-api/multiply")
                .build();

        String responseBody;
        try (Response response = mockClient.newCall(request).execute()) {
            assertEquals(200, response.code());
            assertNotNull(response.body());
            responseBody = response.body().string();
        }

        // Explicitly validate the generated mock response against the schema.
        // Include required query parameters x and y in the path for validation —
        // the schema requires them, even though we omit them from the mock URL.
        ValidationRequest validationRequest = ValidationRequest.builder()
                .method("GET")
                .path("/multiply?x=4&y=7")
                .headers(Map.of())
                .build();

        ValidationResponse validationResponse = ValidationResponse.builder()
                .statusCode(200)
                .header("content-type", "application/json")
                .body(responseBody)
                .build();

        ValidationResult result = validator.validate(validationRequest, validationResponse);

        assertTrue(result.isValid(),
                "Mock response should be valid against schema. Errors: " + result.getErrors());
    }
}

package demo.consumer3;

import io.github.sahina.sdk.ContractValidator;
import io.github.sahina.sdk.ValidationResult;
import io.github.sahina.sdk.adapters.AdapterConfig;
import io.github.sahina.sdk.adapters.CapturedInteraction;
import io.github.sahina.sdk.adapters.OkHttpContractAdapter;

import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;

import org.junit.jupiter.api.AfterAll;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

/**
 * HTTP ADAPTER APPROACH
 * ---------------------
 * Wraps OkHttpClient with OkHttpContractAdapter for automatic validation on every request.
 *
 * Prerequisites:
 * - CVT server running
 * - Producer service running
 *
 * When to use:
 * - When you want transparent validation without modifying each call site
 * - Integration testing with automatic contract enforcement
 */
public class AdapterValidationTest {

    private static final String CVT_SERVER_ADDR =
            System.getenv().getOrDefault("CVT_SERVER_ADDR", "localhost:9550");
    private static final String PRODUCER_URL =
            System.getenv().getOrDefault("PRODUCER_URL", "http://localhost:10001");
    private static final String SCHEMA_PATH =
            System.getenv().getOrDefault("SCHEMA_PATH", "../producer/calculator-api.json");
    private static final String SCHEMA_ID = "calculator-api";

    private static ContractValidator validator;
    private static OkHttpContractAdapter adapter;
    private static OkHttpClient httpClient;

    @BeforeAll
    static void setUp() throws Exception {
        validator = new ContractValidator(CVT_SERVER_ADDR);
        validator.registerSchema(SCHEMA_ID, SCHEMA_PATH);

        adapter = new OkHttpContractAdapter(validator);
        adapter.withConfig(AdapterConfig.builder().autoValidate(true).build());

        httpClient = new OkHttpClient.Builder()
                .addInterceptor(adapter)
                .build();
    }

    @AfterAll
    static void tearDown() throws Exception {
        if (validator != null) {
            validator.close();
        }
    }

    @BeforeEach
    void clearInteractions() {
        if (adapter != null) {
            adapter.clearInteractions();
        }
    }

    @Test
    void shouldAutoValidateMultiplyOperation() throws Exception {
        Request request = new Request.Builder()
                .url(PRODUCER_URL + "/multiply?x=6&y=7")
                .build();

        try (Response response = httpClient.newCall(request).execute()) {
            assertEquals(200, response.code(), "Expected HTTP 200 for multiply");
            assertNotNull(response.body(), "Response body should not be null");
            response.body().string(); // consume body
        }

        List<CapturedInteraction> interactions = adapter.getInteractions();
        assertEquals(1, interactions.size(),
                "Expected exactly 1 captured interaction after multiply request");

        ValidationResult validationResult = interactions.get(0).getValidationResult();
        assertNotNull(validationResult, "Validation result should not be null");
        assertTrue(validationResult.isValid(),
                "Multiply interaction should be valid. Errors: " + validationResult.getErrors());
    }

    @Test
    void shouldAutoValidateDivideOperation() throws Exception {
        Request request = new Request.Builder()
                .url(PRODUCER_URL + "/divide?x=20&y=4")
                .build();

        String body;
        try (Response response = httpClient.newCall(request).execute()) {
            assertEquals(200, response.code(), "Expected HTTP 200 for divide");
            assertNotNull(response.body(), "Response body should not be null");
            body = response.body().string();
        }

        assertTrue(body.contains("5") || body.contains("result"),
                "Response body should contain result=5, got: " + body);

        List<CapturedInteraction> interactions = adapter.getInteractions();
        assertEquals(1, interactions.size(),
                "Expected exactly 1 captured interaction after divide request");

        ValidationResult validationResult = interactions.get(0).getValidationResult();
        assertNotNull(validationResult, "Validation result should not be null");
        assertTrue(validationResult.isValid(),
                "Divide interaction should be valid. Errors: " + validationResult.getErrors());
    }

    @Test
    void shouldCaptureMultipleInteractions() throws Exception {
        // Make multiply request
        Request multiplyRequest = new Request.Builder()
                .url(PRODUCER_URL + "/multiply?x=6&y=7")
                .build();
        try (Response response = httpClient.newCall(multiplyRequest).execute()) {
            assertNotNull(response.body());
            response.body().string();
        }

        // Make divide request
        Request divideRequest = new Request.Builder()
                .url(PRODUCER_URL + "/divide?x=20&y=4")
                .build();
        try (Response response = httpClient.newCall(divideRequest).execute()) {
            assertNotNull(response.body());
            response.body().string();
        }

        List<CapturedInteraction> interactions = adapter.getInteractions();
        assertEquals(2, interactions.size(),
                "Expected 2 captured interactions for multiply and divide");

        for (CapturedInteraction interaction : interactions) {
            ValidationResult validationResult = interaction.getValidationResult();
            assertNotNull(validationResult, "Each interaction should have a validation result");
            assertTrue(validationResult.isValid(),
                    "All interactions should be valid. Errors: " + validationResult.getErrors());
        }
    }
}

package demo.consumer3;

import io.github.sahina.sdk.ContractValidator;
import io.github.sahina.sdk.ConsumerInfo;
import io.github.sahina.sdk.CanIDeployResult;
import io.github.sahina.sdk.adapters.CapturedInteraction;
import io.github.sahina.sdk.AutoRegisterConfig;
import io.github.sahina.sdk.AutoRegisterUtils;
import io.github.sahina.sdk.RegisterConsumerOptions;
import io.github.sahina.sdk.EndpointUsage;
import io.github.sahina.sdk.adapters.MockInterceptor;

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
 * CONSUMER REGISTRATION
 * ---------------------
 * Registers endpoints used by consumer-3 for breaking change detection.
 *
 * Prerequisites:
 * - CVT server running (for schema registration and consumer operations)
 * - No producer service needed (uses MockInterceptor for interaction capture)
 *
 * When to use:
 * - Setting up breaking change detection
 * - Verifying can-i-deploy checks before deployments
 * - Documenting which endpoints a consumer depends on
 */
public class RegistrationTest {

    private static final String CVT_SERVER_ADDR =
            System.getenv().getOrDefault("CVT_SERVER_ADDR", "localhost:9550");
    private static final String SCHEMA_PATH =
            System.getenv().getOrDefault("SCHEMA_PATH", "../producer/calculator-api.json");
    private static final String SCHEMA_ID = "calculator-api";

    private static final String CONSUMER_ID = "consumer-3";
    private static final String CONSUMER_VERSION = "1.0.0";
    private static final String ENVIRONMENT =
            System.getenv().getOrDefault("CVT_ENVIRONMENT", "demo");

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

    private void makeMockRequests() throws Exception {
        Request multiplyRequest = new Request.Builder()
                .url("http://calculator-api/multiply")
                .build();
        try (Response response = mockClient.newCall(multiplyRequest).execute()) {
            assertNotNull(response.body());
            response.body().string();
        }

        Request divideRequest = new Request.Builder()
                .url("http://calculator-api/divide")
                .build();
        try (Response response = mockClient.newCall(divideRequest).execute()) {
            assertNotNull(response.body());
            response.body().string();
        }
    }

    @Test
    void shouldCaptureInteractionsForAutoRegistration() throws Exception {
        makeMockRequests();

        List<CapturedInteraction> interactions = mockInterceptor.getInteractions();
        assertEquals(2, interactions.size(), "Expected 2 interactions for multiply and divide");

        AutoRegisterConfig config = AutoRegisterConfig.builder()
                .consumerId(CONSUMER_ID)
                .consumerVersion(CONSUMER_VERSION)
                .environment(ENVIRONMENT)
                .schemaId(SCHEMA_ID)
                .schemaVersion("1.0.0")
                .build();

        AutoRegisterUtils.BuildResult buildResult =
                validator.buildConsumerFromInteractions(interactions, config);

        assertNotNull(buildResult, "Build result should not be null");
        assertFalse(buildResult.hasError(),
                "Build should succeed without errors. Error: " + buildResult.getError());

        RegisterConsumerOptions options = buildResult.getOptions();
        assertNotNull(options, "Options should not be null");
        assertEquals(CONSUMER_ID, options.getConsumerId(),
                "Consumer ID should match");
        assertEquals(2, options.getUsedEndpoints().size(),
                "Should have 2 used endpoints (multiply and divide)");
    }

    @Test
    void shouldRegisterConsumerFromInteractions() throws Exception {
        makeMockRequests();

        List<CapturedInteraction> interactions = mockInterceptor.getInteractions();
        assertEquals(2, interactions.size(), "Expected 2 interactions");

        AutoRegisterConfig config = AutoRegisterConfig.builder()
                .consumerId(CONSUMER_ID)
                .consumerVersion(CONSUMER_VERSION)
                .environment(ENVIRONMENT)
                .schemaId(SCHEMA_ID)
                .schemaVersion("1.0.0")
                .build();

        ConsumerInfo consumer = validator.registerConsumerFromInteractions(interactions, config);

        assertNotNull(consumer, "Consumer info should not be null");
        assertEquals(CONSUMER_ID, consumer.getConsumerId(),
                "Registered consumer ID should match");
    }

    @Test
    void shouldRegisterConsumerWithExplicitEndpoints() throws Exception {
        RegisterConsumerOptions options = RegisterConsumerOptions.builder()
                .consumerId(CONSUMER_ID)
                .consumerVersion(CONSUMER_VERSION)
                .schemaId(SCHEMA_ID)
                .schemaVersion("1.0.0")
                .environment(ENVIRONMENT)
                .usedEndpoints(List.of(
                        new EndpointUsage("GET", "/multiply", List.of("result")),
                        new EndpointUsage("GET", "/divide", List.of("result"))
                ))
                .build();

        ConsumerInfo consumer = validator.registerConsumer(options);

        assertNotNull(consumer, "Consumer info should not be null");
        assertEquals(CONSUMER_ID, consumer.getConsumerId(),
                "Registered consumer ID should match");
    }

    @Test
    void shouldListRegisteredConsumers() throws Exception {
        // Register the consumer first
        RegisterConsumerOptions options = RegisterConsumerOptions.builder()
                .consumerId(CONSUMER_ID)
                .consumerVersion(CONSUMER_VERSION)
                .schemaId(SCHEMA_ID)
                .schemaVersion("1.0.0")
                .environment(ENVIRONMENT)
                .usedEndpoints(List.of(
                        new EndpointUsage("GET", "/multiply", List.of("result")),
                        new EndpointUsage("GET", "/divide", List.of("result"))
                ))
                .build();
        validator.registerConsumer(options);

        List<ConsumerInfo> consumers = validator.listConsumers(SCHEMA_ID, ENVIRONMENT);

        assertNotNull(consumers, "Consumer list should not be null");
        boolean found = consumers.stream()
                .anyMatch(c -> CONSUMER_ID.equals(c.getConsumerId()));
        assertTrue(found,
                "Registered consumer '" + CONSUMER_ID + "' should appear in consumer list");
    }

    @Test
    void shouldCheckDeploymentSafetyWithCanIDeploy() throws Exception {
        // Register the consumer first
        RegisterConsumerOptions options = RegisterConsumerOptions.builder()
                .consumerId(CONSUMER_ID)
                .consumerVersion(CONSUMER_VERSION)
                .schemaId(SCHEMA_ID)
                .schemaVersion("1.0.0")
                .environment(ENVIRONMENT)
                .usedEndpoints(List.of(
                        new EndpointUsage("GET", "/multiply", List.of("result")),
                        new EndpointUsage("GET", "/divide", List.of("result"))
                ))
                .build();
        validator.registerConsumer(options);

        CanIDeployResult result = validator.canIDeploy(SCHEMA_ID, "1.0.0", ENVIRONMENT);

        assertNotNull(result, "Can-i-deploy result should not be null");
    }
}

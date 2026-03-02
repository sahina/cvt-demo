# Consumer-3 (Java) and Consumer-4 (Go) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add consumer-3 (Java 21/Maven, multiply+divide) and consumer-4 (Go 1.25, add+subtract) to the CVT demo project with full parity: CLI, 4 test types each, Docker, Makefile targets, CI jobs, and updated docs.

**Architecture:** Each consumer follows the identical pattern as consumer-1 and consumer-2: a CLI entry point, four test files (mock, manual, adapter, registration), a multi-stage Dockerfile, Makefile targets for operations and tests, and a GitHub Actions job in both workflow files. Consumer-4 reuses the same Go CVT SDK already used by the producer (`github.com/sahina/cvt/sdks/go v0.3.0`). Consumer-3 uses the Java SDK (`io.github.sahina:cvt-sdk:0.3.0`).

**Tech Stack:**
- Consumer-3: Java 21, Maven, `io.github.sahina:cvt-sdk:0.3.0`, JUnit 5, Java built-in `HttpClient`
- Consumer-4: Go 1.25, `github.com/sahina/cvt/sdks/go v0.3.0`, stdlib `net/http` + `testing`
- Shared: `ghcr.io/sahina/cvt:0.3.0` server, `producer/calculator-api.json` schema

---

## Pre-flight: Understand the CVT SDK APIs

Before writing any consumer code, discover the exact method signatures from the published SDKs. The Java SDK API is inferred from Node.js/Python patterns and must be verified.

### Task 0: Discover Java SDK API

**Files:**
- No files to create — this is research

**Step 1: Pull the Java SDK JAR and inspect it**

```bash
# Create a temp Maven project to pull and inspect the SDK
mkdir /tmp/cvt-java-inspect && cd /tmp/cvt-java-inspect

# Fetch the JAR (Maven Central / GitHub Packages)
mvn dependency:get \
  -Dartifact=io.github.sahina:cvt-sdk:0.3.0 \
  -Ddest=/tmp/cvt-sdk.jar

# List all class names in the JAR
jar tf /tmp/cvt-sdk.jar | grep '\.class' | sed 's|/|.|g' | sed 's|\.class||'
```

Expected: You'll see class names like `io.github.sahina.cvt.ContractValidator`, `io.github.sahina.cvt.adapters.MockHttpClient`, etc.

**Step 2: Check for a README or javadoc**

```bash
# Look for source or javadoc JAR
mvn dependency:get \
  -Dartifact=io.github.sahina:cvt-sdk:0.3.0:jar:sources \
  -Ddest=/tmp/cvt-sdk-sources.jar

jar tf /tmp/cvt-sdk-sources.jar | head -50
```

**Step 3: Note the exact class and method names**

You need to find equivalents of:
- `ContractValidator` constructor (takes CVT server address string)
- `registerSchema(schemaId, schemaPath)` method
- `validate(request, response)` method
- Mock adapter class name and constructor
- HTTP adapter class name (for auto-validation)
- `registerConsumer(options)` method
- `buildConsumerFromInteractions(interactions, config)` method
- `registerConsumerFromInteractions(interactions, config)` method
- `canIDeploy(schemaId, version, environment)` method
- `listConsumers(schemaId, environment)` method

**Step 4: Discover Go SDK consumer APIs**

```bash
# From the existing producer go.sum, find the SDK module path
cd /path/to/cvt-demo/producer
go doc github.com/sahina/cvt/sdks/go/cvt
go doc github.com/sahina/cvt/sdks/go/cvt.Validator
```

Look for:
- `RegisterConsumer` or similar on `*cvt.Validator`
- Mock client package (e.g., `github.com/sahina/cvt/sdks/go/cvt/consumer`)
- HTTP adapter for `net/http` (e.g., `github.com/sahina/cvt/sdks/go/cvt/consumer/adapters`)

> **NOTE FOR ALL SUBSEQUENT TASKS:** Adjust class/method names based on what you discovered here. The code in this plan uses the most likely names based on Node.js/Python/Go producer patterns. Verify before using.

---

## Phase 1: Consumer-3 (Java)

### Task 1: Set Up Maven Project Structure

**Files:**
- Create: `consumer-3/pom.xml`
- Create: `consumer-3/src/main/java/demo/consumer3/.gitkeep`
- Create: `consumer-3/src/test/java/demo/consumer3/.gitkeep`

**Step 1: Create directory structure**

```bash
mkdir -p consumer-3/src/main/java/demo/consumer3
mkdir -p consumer-3/src/test/java/demo/consumer3
```

**Step 2: Create `consumer-3/pom.xml`**

```xml
<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>demo</groupId>
    <artifactId>consumer-3</artifactId>
    <version>1.0.0</version>
    <packaging>jar</packaging>

    <properties>
        <maven.compiler.source>21</maven.compiler.source>
        <maven.compiler.target>21</maven.compiler.target>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <junit.version>5.10.2</junit.version>
    </properties>

    <dependencies>
        <!-- CVT SDK for contract validation -->
        <dependency>
            <groupId>io.github.sahina</groupId>
            <artifactId>cvt-sdk</artifactId>
            <version>0.3.0</version>
        </dependency>

        <!-- JUnit 5 for testing -->
        <dependency>
            <groupId>org.junit.jupiter</groupId>
            <artifactId>junit-jupiter</artifactId>
            <version>${junit.version}</version>
            <scope>test</scope>
        </dependency>
    </dependencies>

    <build>
        <plugins>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-surefire-plugin</artifactId>
                <version>3.2.5</version>
                <configuration>
                    <includes>
                        <include>**/*Test.java</include>
                    </includes>
                </configuration>
            </plugin>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-jar-plugin</artifactId>
                <version>3.3.0</version>
                <configuration>
                    <archive>
                        <manifest>
                            <mainClass>demo.consumer3.Main</mainClass>
                        </manifest>
                    </archive>
                </configuration>
            </plugin>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-shade-plugin</artifactId>
                <version>3.5.2</version>
                <executions>
                    <execution>
                        <phase>package</phase>
                        <goals>
                            <goal>shade</goal>
                        </goals>
                        <configuration>
                            <finalName>consumer3</finalName>
                            <createDependencyReducedPom>false</createDependencyReducedPom>
                        </configuration>
                    </execution>
                </executions>
            </plugin>
        </plugins>
    </build>
</project>
```

**Step 3: Verify Maven resolves the SDK**

```bash
cd consumer-3
mvn dependency:resolve -q
```

Expected: `BUILD SUCCESS` with `io.github.sahina:cvt-sdk:jar:0.3.0` in the dependency tree.

**Step 4: Commit**

```bash
git add consumer-3/pom.xml consumer-3/src/
git commit -m "feat(consumer-3): scaffold Java Maven project structure"
```

---

### Task 2: Consumer-3 — MockValidationTest

**Files:**
- Create: `consumer-3/src/test/java/demo/consumer3/MockValidationTest.java`

> **TDD NOTE:** Write the test first, run it to see it fail with "class not found" or SDK errors, then adjust based on actual SDK API.

**Step 1: Write `MockValidationTest.java`**

```java
package demo.consumer3;

import io.github.sahina.cvt.ContractValidator;
// VERIFY: import the mock adapter — likely one of:
//   io.github.sahina.cvt.adapters.MockHttpClient
//   io.github.sahina.cvt.adapters.MockClient
//   io.github.sahina.cvt.adapters.CvtMockAdapter
import io.github.sahina.cvt.adapters.MockHttpClient;

import org.junit.jupiter.api.*;
import static org.junit.jupiter.api.Assertions.*;

import java.net.http.HttpResponse;
import java.util.List;

/**
 * MOCK ADAPTER APPROACH
 * ---------------------
 * Generates mock responses based on the OpenAPI schema without making real HTTP calls.
 *
 * Prerequisites:
 * - CVT server running (for schema validation)
 * - No producer service needed
 */
@TestMethodOrder(MethodOrderer.OrderAnnotation.class)
class MockValidationTest {

    static final String CVT_SERVER_ADDR = System.getenv().getOrDefault("CVT_SERVER_ADDR", "localhost:9550");
    static final String SCHEMA_PATH = System.getenv().getOrDefault("SCHEMA_PATH",
            "../../producer/calculator-api.json");

    static ContractValidator validator;
    // VERIFY: adjust type to match actual mock adapter class
    static MockHttpClient mockClient;

    @BeforeAll
    static void setUp() throws Exception {
        validator = new ContractValidator(CVT_SERVER_ADDR);
        validator.registerSchema("calculator-api", SCHEMA_PATH);

        // VERIFY: adjust constructor args to match actual SDK
        mockClient = new MockHttpClient(validator);
    }

    @AfterAll
    static void tearDown() throws Exception {
        if (mockClient != null) mockClient.close();
        if (validator != null) validator.close();
    }

    @BeforeEach
    void clearInteractions() {
        mockClient.clearInteractions();
    }

    @Test
    @Order(1)
    void shouldGenerateValidMockResponseForMultiply() throws Exception {
        // VERIFY: adjust fetch/get method name and URL handling
        HttpResponse<String> response = mockClient.fetch("http://calculator-api/multiply?x=4&y=7");

        assertEquals(200, response.statusCode());
        String body = response.body();
        assertTrue(body.contains("result"), "Response should contain 'result' field, got: " + body);
    }

    @Test
    @Order(2)
    void shouldGenerateValidMockResponseForDivide() throws Exception {
        HttpResponse<String> response = mockClient.fetch("http://calculator-api/divide?x=10&y=2");

        assertEquals(200, response.statusCode());
        String body = response.body();
        assertTrue(body.contains("result"), "Response should contain 'result' field, got: " + body);
    }

    @Test
    @Order(3)
    void shouldCaptureMockInteractions() throws Exception {
        mockClient.fetch("http://calculator-api/multiply?x=5&y=6");

        // VERIFY: adjust getInteractions() return type
        List<?> interactions = mockClient.getInteractions();
        assertEquals(1, interactions.size());
    }

    @Test
    @Order(4)
    void shouldCaptureAllConsumer3Endpoints() throws Exception {
        mockClient.fetch("http://calculator-api/multiply?x=4&y=7");
        mockClient.fetch("http://calculator-api/divide?x=10&y=2");

        List<?> interactions = mockClient.getInteractions();
        assertEquals(2, interactions.size());
    }

    @Test
    @Order(5)
    void mockResponseShouldValidateAgainstSchema() throws Exception {
        HttpResponse<String> response = mockClient.fetch("http://calculator-api/multiply?x=4&y=7");

        // VERIFY: adjust ValidationRequest/Response class names
        var request = new io.github.sahina.cvt.ValidationRequest("GET", "/multiply?x=4&y=7", null);
        var validationResp = new io.github.sahina.cvt.ValidationResponse(200,
                java.util.Map.of("content-type", "application/json"),
                response.body());

        var result = validator.validate(request, validationResp);
        assertTrue(result.isValid(), "Mock response should validate against schema");
    }
}
```

**Step 2: Run test to verify it compiles and connects**

```bash
cd consumer-3
CVT_SERVER_ADDR=localhost:9550 \
SCHEMA_PATH=../producer/calculator-api.json \
mvn test -Dtest="MockValidationTest" -pl .
```

Expected: Tests run. If import errors occur, adjust import paths based on Task 0 SDK discovery. If all green: move on.

**Step 3: Fix any import/API mismatches found, re-run until green**

```bash
mvn test -Dtest="MockValidationTest"
```

Expected: `BUILD SUCCESS`, all 5 tests PASS.

**Step 4: Commit**

```bash
git add consumer-3/src/test/java/demo/consumer3/MockValidationTest.java
git commit -m "feat(consumer-3): add mock validation tests"
```

---

### Task 3: Consumer-3 — ManualValidationTest

**Files:**
- Create: `consumer-3/src/test/java/demo/consumer3/ManualValidationTest.java`

**Step 1: Write `ManualValidationTest.java`**

```java
package demo.consumer3;

import io.github.sahina.cvt.ContractValidator;
import io.github.sahina.cvt.ValidationRequest;
import io.github.sahina.cvt.ValidationResponse;
import io.github.sahina.cvt.ValidationResult;

import org.junit.jupiter.api.*;
import static org.junit.jupiter.api.Assertions.*;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.Map;

/**
 * MANUAL VALIDATION APPROACH
 * --------------------------
 * Explicitly validates request/response pairs by calling validator.validate() directly.
 *
 * Prerequisites:
 * - CVT server running
 * - Producer service running
 */
class ManualValidationTest {

    static final String CVT_SERVER_ADDR = System.getenv().getOrDefault("CVT_SERVER_ADDR", "localhost:9550");
    static final String PRODUCER_URL = System.getenv().getOrDefault("PRODUCER_URL", "http://localhost:10001");
    static final String SCHEMA_PATH = System.getenv().getOrDefault("SCHEMA_PATH",
            "../../producer/calculator-api.json");

    static ContractValidator validator;
    static HttpClient httpClient;

    @BeforeAll
    static void setUp() throws Exception {
        validator = new ContractValidator(CVT_SERVER_ADDR);
        validator.registerSchema("calculator-api", SCHEMA_PATH);
        httpClient = HttpClient.newHttpClient();
    }

    @AfterAll
    static void tearDown() throws Exception {
        if (validator != null) validator.close();
    }

    @Test
    void shouldValidateSuccessfulMultiplyOperation() throws Exception {
        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create(PRODUCER_URL + "/multiply?x=4&y=7"))
                .GET()
                .build();

        HttpResponse<String> resp = httpClient.send(req, HttpResponse.BodyHandlers.ofString());

        assertEquals(200, resp.statusCode());

        // VERIFY: adjust ValidationRequest/ValidationResponse constructor to match SDK
        ValidationRequest request = new ValidationRequest("GET", "/multiply?x=4&y=7", Map.of());
        ValidationResponse response = new ValidationResponse(
                resp.statusCode(),
                Map.of("content-type", "application/json"),
                resp.body()
        );

        ValidationResult result = validator.validate(request, response);

        assertTrue(result.isValid(), "Multiply response should be valid");
        assertTrue(result.getErrors().isEmpty());

        // Parse body manually to verify the math
        assertTrue(resp.body().contains("\"result\":28") || resp.body().contains("\"result\": 28"),
                "4 * 7 should equal 28, got: " + resp.body());
    }

    @Test
    void shouldValidateSuccessfulDivideOperation() throws Exception {
        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create(PRODUCER_URL + "/divide?x=10&y=2"))
                .GET()
                .build();

        HttpResponse<String> resp = httpClient.send(req, HttpResponse.BodyHandlers.ofString());

        assertEquals(200, resp.statusCode());

        ValidationRequest request = new ValidationRequest("GET", "/divide?x=10&y=2", Map.of());
        ValidationResponse response = new ValidationResponse(
                resp.statusCode(),
                Map.of("content-type", "application/json"),
                resp.body()
        );

        ValidationResult result = validator.validate(request, response);

        assertTrue(result.isValid());
        assertTrue(resp.body().contains("\"result\":5") || resp.body().contains("\"result\": 5"),
                "10 / 2 should equal 5, got: " + resp.body());
    }

    @Test
    void shouldDetectMissingResultField() throws Exception {
        ValidationRequest request = new ValidationRequest("GET", "/multiply?x=4&y=7", Map.of());
        ValidationResponse response = new ValidationResponse(
                200,
                Map.of("content-type", "application/json"),
                "{\"product\": 28}"
        );

        ValidationResult result = validator.validate(request, response);

        assertFalse(result.isValid(), "Response with wrong field name should fail validation");
        assertFalse(result.getErrors().isEmpty());
    }

    @Test
    void shouldValidateDivideByZeroErrorResponse() throws Exception {
        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create(PRODUCER_URL + "/divide?x=10&y=0"))
                .GET()
                .build();

        HttpResponse<String> resp = httpClient.send(req, HttpResponse.BodyHandlers.ofString());

        assertEquals(400, resp.statusCode());

        ValidationRequest request = new ValidationRequest("GET", "/divide?x=10&y=0", Map.of());
        ValidationResponse response = new ValidationResponse(
                resp.statusCode(),
                Map.of("content-type", "application/json"),
                resp.body()
        );

        ValidationResult result = validator.validate(request, response);
        assertTrue(result.isValid(), "400 error response should match error schema");
    }
}
```

**Step 2: Run with producer running**

```bash
# In another terminal: make up
CVT_SERVER_ADDR=localhost:9550 \
PRODUCER_URL=http://localhost:10001 \
SCHEMA_PATH=../producer/calculator-api.json \
mvn test -Dtest="ManualValidationTest"
```

Expected: `BUILD SUCCESS`, 4 tests PASS.

**Step 3: Commit**

```bash
git add consumer-3/src/test/java/demo/consumer3/ManualValidationTest.java
git commit -m "feat(consumer-3): add manual validation tests"
```

---

### Task 4: Consumer-3 — AdapterValidationTest

**Files:**
- Create: `consumer-3/src/test/java/demo/consumer3/AdapterValidationTest.java`

**Step 1: Write `AdapterValidationTest.java`**

```java
package demo.consumer3;

import io.github.sahina.cvt.ContractValidator;
// VERIFY: import the HTTP adapter class — likely one of:
//   io.github.sahina.cvt.adapters.ContractValidatingHttpClient
//   io.github.sahina.cvt.adapters.ValidatingHttpClient
//   io.github.sahina.cvt.adapters.HttpClientAdapter
import io.github.sahina.cvt.adapters.ContractValidatingHttpClient;

import org.junit.jupiter.api.*;
import static org.junit.jupiter.api.Assertions.*;

import java.net.URI;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.util.List;

/**
 * HTTP ADAPTER APPROACH
 * ---------------------
 * Wraps the HTTP client to automatically validate all requests/responses.
 *
 * Prerequisites:
 * - CVT server running
 * - Producer service running
 */
class AdapterValidationTest {

    static final String CVT_SERVER_ADDR = System.getenv().getOrDefault("CVT_SERVER_ADDR", "localhost:9550");
    static final String PRODUCER_URL = System.getenv().getOrDefault("PRODUCER_URL", "http://localhost:10001");
    static final String SCHEMA_PATH = System.getenv().getOrDefault("SCHEMA_PATH",
            "../../producer/calculator-api.json");

    static ContractValidator validator;
    // VERIFY: adjust type to match actual adapter class
    static ContractValidatingHttpClient validatingClient;

    @BeforeAll
    static void setUp() throws Exception {
        validator = new ContractValidator(CVT_SERVER_ADDR);
        validator.registerSchema("calculator-api", SCHEMA_PATH);

        // VERIFY: adjust constructor/builder to match actual SDK
        validatingClient = new ContractValidatingHttpClient(validator, true /* autoValidate */);
    }

    @AfterAll
    static void tearDown() throws Exception {
        if (validator != null) validator.close();
    }

    @BeforeEach
    void clearInteractions() {
        validatingClient.clearInteractions();
    }

    @Test
    void shouldAutoValidateMultiplyOperation() throws Exception {
        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create(PRODUCER_URL + "/multiply?x=6&y=7"))
                .GET()
                .build();

        HttpResponse<String> response = validatingClient.send(req, HttpResponse.BodyHandlers.ofString());

        assertEquals(200, response.statusCode());

        List<?> interactions = validatingClient.getInteractions();
        assertFalse(interactions.isEmpty());
        // VERIFY: check interaction result field name
        Object lastInteraction = interactions.get(interactions.size() - 1);
        // e.g.: assertTrue(((Interaction) lastInteraction).getValidationResult().isValid());
    }

    @Test
    void shouldAutoValidateDivideOperation() throws Exception {
        HttpRequest req = HttpRequest.newBuilder()
                .uri(URI.create(PRODUCER_URL + "/divide?x=20&y=4"))
                .GET()
                .build();

        HttpResponse<String> response = validatingClient.send(req, HttpResponse.BodyHandlers.ofString());

        assertEquals(200, response.statusCode());
        assertTrue(response.body().contains("\"result\":5") || response.body().contains("\"result\": 5"),
                "20 / 4 = 5, got: " + response.body());

        List<?> interactions = validatingClient.getInteractions();
        assertEquals(1, interactions.size());
    }

    @Test
    void shouldCaptureMultipleInteractions() throws Exception {
        validatingClient.send(
                HttpRequest.newBuilder().uri(URI.create(PRODUCER_URL + "/multiply?x=3&y=4")).GET().build(),
                HttpResponse.BodyHandlers.ofString());
        validatingClient.send(
                HttpRequest.newBuilder().uri(URI.create(PRODUCER_URL + "/divide?x=8&y=2")).GET().build(),
                HttpResponse.BodyHandlers.ofString());

        List<?> interactions = validatingClient.getInteractions();
        assertEquals(2, interactions.size());
    }
}
```

**Step 2: Run tests**

```bash
CVT_SERVER_ADDR=localhost:9550 \
PRODUCER_URL=http://localhost:10001 \
SCHEMA_PATH=../producer/calculator-api.json \
mvn test -Dtest="AdapterValidationTest"
```

Expected: `BUILD SUCCESS`, 3 tests PASS.

**Step 3: Commit**

```bash
git add consumer-3/src/test/java/demo/consumer3/AdapterValidationTest.java
git commit -m "feat(consumer-3): add adapter validation tests"
```

---

### Task 5: Consumer-3 — RegistrationTest

**Files:**
- Create: `consumer-3/src/test/java/demo/consumer3/RegistrationTest.java`

**Step 1: Write `RegistrationTest.java`**

```java
package demo.consumer3;

import io.github.sahina.cvt.ContractValidator;
import io.github.sahina.cvt.adapters.MockHttpClient;
// VERIFY: import registration options classes — likely:
//   io.github.sahina.cvt.RegisterConsumerOptions
//   io.github.sahina.cvt.AutoRegisterConfig
//   io.github.sahina.cvt.EndpointUsage
import io.github.sahina.cvt.RegisterConsumerOptions;
import io.github.sahina.cvt.AutoRegisterConfig;
import io.github.sahina.cvt.EndpointUsage;

import org.junit.jupiter.api.*;
import static org.junit.jupiter.api.Assertions.*;

import java.util.List;
import java.util.Map;

/**
 * CONSUMER REGISTRATION
 * ---------------------
 * Registers which API endpoints this consumer uses so breaking changes can be detected.
 *
 * Prerequisites:
 * - CVT server running
 */
class RegistrationTest {

    static final String CVT_SERVER_ADDR = System.getenv().getOrDefault("CVT_SERVER_ADDR", "localhost:9550");
    static final String SCHEMA_PATH = System.getenv().getOrDefault("SCHEMA_PATH",
            "../../producer/calculator-api.json");
    static final String CONSUMER_ID = "consumer-3";
    static final String CONSUMER_VERSION = "1.0.0";
    static final String ENVIRONMENT = System.getenv().getOrDefault("CVT_ENVIRONMENT", "demo");

    static ContractValidator validator;
    static MockHttpClient mockClient;

    @BeforeAll
    static void setUp() throws Exception {
        validator = new ContractValidator(CVT_SERVER_ADDR);
        validator.registerSchema("calculator-api", SCHEMA_PATH);
        mockClient = new MockHttpClient(validator);
    }

    @AfterAll
    static void tearDown() throws Exception {
        if (mockClient != null) mockClient.close();
        if (validator != null) validator.close();
    }

    @BeforeEach
    void clearInteractions() {
        mockClient.clearInteractions();
    }

    @Test
    void shouldCaptureInteractionsForAutoRegistration() throws Exception {
        mockClient.fetch("http://calculator-api/multiply?x=4&y=7");
        mockClient.fetch("http://calculator-api/divide?x=10&y=2");

        List<?> interactions = mockClient.getInteractions();
        assertEquals(2, interactions.size());

        // VERIFY: adjust AutoRegisterConfig constructor/builder to match SDK
        AutoRegisterConfig config = new AutoRegisterConfig(
                CONSUMER_ID, CONSUMER_VERSION, ENVIRONMENT, "1.0.0", "calculator-api");

        var opts = validator.buildConsumerFromInteractions(interactions, config);

        assertNotNull(opts);
        assertEquals(CONSUMER_ID, opts.getConsumerId());
        assertNotNull(opts.getUsedEndpoints());
        assertEquals(2, opts.getUsedEndpoints().size());
    }

    @Test
    void shouldRegisterConsumerFromInteractions() throws Exception {
        mockClient.fetch("http://calculator-api/multiply?x=1&y=2");
        mockClient.fetch("http://calculator-api/divide?x=4&y=2");

        List<?> interactions = mockClient.getInteractions();

        AutoRegisterConfig config = new AutoRegisterConfig(
                CONSUMER_ID, CONSUMER_VERSION, ENVIRONMENT, "1.0.0", "calculator-api");

        var consumer = validator.registerConsumerFromInteractions(interactions, config);

        assertNotNull(consumer);
        // VERIFY: adjust getter name to match SDK
        assertEquals(CONSUMER_ID, consumer.getConsumerId());
    }

    @Test
    void shouldRegisterConsumerWithExplicitEndpoints() throws Exception {
        // VERIFY: adjust RegisterConsumerOptions and EndpointUsage constructors
        RegisterConsumerOptions options = new RegisterConsumerOptions(
                CONSUMER_ID, CONSUMER_VERSION,
                "calculator-api", "1.0.0", ENVIRONMENT,
                List.of(
                        new EndpointUsage("GET", "/multiply", List.of("result")),
                        new EndpointUsage("GET", "/divide", List.of("result"))
                )
        );

        var consumer = validator.registerConsumer(options);

        assertNotNull(consumer);
        assertEquals(CONSUMER_ID, consumer.getConsumerId());
    }

    @Test
    void shouldListRegisteredConsumers() throws Exception {
        // Register first to ensure at least one consumer exists
        RegisterConsumerOptions options = new RegisterConsumerOptions(
                CONSUMER_ID, CONSUMER_VERSION,
                "calculator-api", "1.0.0", ENVIRONMENT,
                List.of(
                        new EndpointUsage("GET", "/multiply", List.of("result")),
                        new EndpointUsage("GET", "/divide", List.of("result"))
                )
        );
        validator.registerConsumer(options);

        List<?> consumers = validator.listConsumers("calculator-api", ENVIRONMENT);

        assertNotNull(consumers);
        assertFalse(consumers.isEmpty());
        assertTrue(consumers.stream().anyMatch(c -> CONSUMER_ID.equals(getConsumerId(c))));
    }

    @Test
    void shouldCheckDeploymentSafetyWithCanIDeploy() throws Exception {
        RegisterConsumerOptions options = new RegisterConsumerOptions(
                CONSUMER_ID, CONSUMER_VERSION,
                "calculator-api", "1.0.0", ENVIRONMENT,
                List.of(
                        new EndpointUsage("GET", "/multiply", List.of("result")),
                        new EndpointUsage("GET", "/divide", List.of("result"))
                )
        );
        validator.registerConsumer(options);

        // VERIFY: adjust return type and method name
        var result = validator.canIDeploy("calculator-api", "1.0.0", ENVIRONMENT);

        assertNotNull(result);
        // VERIFY: adjust getter name
        assertNotNull(result.isSafeToDeploy());
    }

    // VERIFY: adjust this helper based on actual consumer return type
    private String getConsumerId(Object consumer) {
        try {
            return (String) consumer.getClass().getMethod("getConsumerId").invoke(consumer);
        } catch (Exception e) {
            return consumer.toString();
        }
    }
}
```

**Step 2: Run tests**

```bash
CVT_SERVER_ADDR=localhost:9550 \
SCHEMA_PATH=../producer/calculator-api.json \
mvn test -Dtest="RegistrationTest"
```

Expected: `BUILD SUCCESS`, 5 tests PASS.

**Step 3: Commit**

```bash
git add consumer-3/src/test/java/demo/consumer3/RegistrationTest.java
git commit -m "feat(consumer-3): add consumer registration tests"
```

---

### Task 6: Consumer-3 — Main.java CLI

**Files:**
- Create: `consumer-3/src/main/java/demo/consumer3/Main.java`

**Step 1: Write `Main.java`**

```java
package demo.consumer3;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

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

    static final String PRODUCER_URL = System.getenv().getOrDefault("PRODUCER_URL", "http://localhost:10001");
    static final String CVT_SERVER_ADDR = System.getenv().getOrDefault("CVT_SERVER_ADDR", "localhost:9550");
    static final String SCHEMA_PATH = System.getenv().getOrDefault("SCHEMA_PATH", "./calculator-api.json");

    public static void main(String[] args) throws Exception {
        if (args.length < 3) {
            printUsage();
            System.exit(1);
        }

        String command = args[0];
        double x, y;
        try {
            x = Double.parseDouble(args[1]);
            y = Double.parseDouble(args[2]);
        } catch (NumberFormatException e) {
            System.err.println("Error: Both arguments must be valid numbers");
            System.exit(1);
            return;
        }

        boolean validate = args.length > 3 && "--validate".equals(args[3]);

        switch (command) {
            case "multiply" -> multiply(x, y, validate);
            case "divide"   -> divide(x, y, validate);
            default -> {
                System.err.println("Error: Unknown command '" + command + "'");
                printUsage();
                System.exit(1);
            }
        }
    }

    static void multiply(double x, double y, boolean validate) throws Exception {
        String path = "/multiply?x=" + x + "&y=" + y;
        String result = callApi(path, validate);
        double value = parseResult(result);
        System.out.printf("%.0f * %.0f = %.0f%n", x, y, value);
    }

    static void divide(double x, double y, boolean validate) throws Exception {
        String path = "/divide?x=" + x + "&y=" + y;
        String result = callApi(path, validate);
        double value = parseResult(result);
        System.out.printf("%.0f / %.0f = %s%n", x, y, formatResult(value));
    }

    static String callApi(String path, boolean validate) throws Exception {
        HttpClient client = HttpClient.newHttpClient();
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(PRODUCER_URL + path))
                .header("Accept", "application/json")
                .GET()
                .build();

        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());

        if (response.statusCode() >= 400) {
            String errorMsg = extractError(response.body());
            System.err.println("Error: " + errorMsg);
            System.exit(1);
        }

        if (validate) {
            runValidation(path, response);
        }

        return response.body();
    }

    static void runValidation(String path, HttpResponse<String> response) {
        try {
            // VERIFY: adjust imports and class names based on SDK discovery in Task 0
            var validator = new io.github.sahina.cvt.ContractValidator(CVT_SERVER_ADDR);
            validator.registerSchema("calculator-api", SCHEMA_PATH);

            var request = new io.github.sahina.cvt.ValidationRequest("GET", path, java.util.Map.of());
            var validationResponse = new io.github.sahina.cvt.ValidationResponse(
                    response.statusCode(),
                    java.util.Map.of("content-type", "application/json"),
                    response.body()
            );

            var result = validator.validate(request, validationResponse);
            validator.close();

            if (!result.isValid()) {
                System.err.println("CVT Validation failed: " + result.getErrors());
                System.exit(1);
            }
            System.err.println("CVT validation passed");
        } catch (Exception e) {
            System.err.println("Warning: CVT validation failed: " + e.getMessage());
        }
    }

    static double parseResult(String json) {
        // Simple JSON parsing without external library
        // Expects: {"result": <number>}
        int idx = json.indexOf("\"result\"");
        if (idx == -1) {
            System.err.println("Error: Unexpected response format: " + json);
            System.exit(1);
        }
        int colon = json.indexOf(":", idx);
        int end = json.indexOf("}", colon);
        String value = json.substring(colon + 1, end).trim().replaceAll("[^0-9.\\-]", "");
        return Double.parseDouble(value);
    }

    static String extractError(String json) {
        int idx = json.indexOf("\"error\"");
        if (idx == -1) return json;
        int colon = json.indexOf(":", idx);
        int start = json.indexOf("\"", colon + 1) + 1;
        int end = json.indexOf("\"", start);
        return json.substring(start, end);
    }

    static String formatResult(double value) {
        if (value == Math.floor(value) && !Double.isInfinite(value)) {
            return String.format("%.0f", value);
        }
        return String.valueOf(value);
    }

    static void printUsage() {
        System.err.println("Usage: java -jar consumer3.jar <command> <x> <y> [--validate]");
        System.err.println("Commands: multiply, divide");
    }
}
```

**Step 2: Build and test locally**

```bash
cd consumer-3
mvn package -DskipTests -q

# Test without validation
java -jar target/consumer3.jar multiply 4 7
# Expected: 4 * 7 = 28

java -jar target/consumer3.jar divide 10 2
# Expected: 10 / 2 = 5

# Test with validation (requires CVT server + producer)
java -jar target/consumer3.jar multiply 4 7 --validate
# Expected: CVT validation passed\n4 * 7 = 28
```

**Step 3: Commit**

```bash
git add consumer-3/src/main/java/demo/consumer3/Main.java
git commit -m "feat(consumer-3): add CLI main entry point"
```

---

### Task 7: Consumer-3 — Dockerfile

**Files:**
- Create: `consumer-3/Dockerfile`

**Step 1: Write `consumer-3/Dockerfile`**

```dockerfile
# Build stage
FROM eclipse-temurin:21 AS builder

WORKDIR /build

# Copy Maven configuration and download dependencies (layer cache)
COPY consumer-3/pom.xml .
RUN apt-get update -qq && apt-get install -y -qq maven && \
    mvn dependency:go-offline -q

# Copy source and build
COPY consumer-3/src ./src
RUN mvn package -DskipTests -q

# Runtime stage
FROM eclipse-temurin:21-jre

WORKDIR /app

# Copy the fat JAR
COPY --from=builder /build/target/consumer3.jar .

# Copy the OpenAPI schema for validation (JSON format required by CVT server)
COPY producer/calculator-api.json ./calculator-api.json

# Set environment variables
ENV PRODUCER_URL=http://producer:10001
ENV CVT_SERVER_ADDR=cvt:9550
ENV SCHEMA_PATH=/app/calculator-api.json

ENTRYPOINT ["java", "-jar", "consumer3.jar"]
```

**Step 2: Build the Docker image from repo root**

```bash
cd /path/to/cvt-demo
docker build -f consumer-3/Dockerfile -t consumer-3-test .
```

Expected: `Successfully built <image-id>` with no errors.

**Step 3: Smoke-test the container (requires producer + CVT running)**

```bash
# make up  (if not already running)
docker run --rm --network cvt-demo-network \
  -e PRODUCER_URL=http://producer:10001 \
  -e CVT_SERVER_ADDR=cvt:9550 \
  -e SCHEMA_PATH=/app/calculator-api.json \
  consumer-3-test multiply 4 7
# Expected: 4 * 7 = 28
```

**Step 4: Commit**

```bash
git add consumer-3/Dockerfile
git commit -m "feat(consumer-3): add multi-stage Dockerfile"
```

---

## Phase 2: Consumer-4 (Go)

### Task 8: Set Up Go Module

**Files:**
- Create: `consumer-4/go.mod`
- Create: `consumer-4/tests/` directory

**Step 1: Initialize Go module**

```bash
mkdir -p consumer-4/tests
cd consumer-4
go mod init github.com/sahina/cvt-demo/consumer-4
```

**Step 2: Add CVT SDK dependency**

```bash
cd consumer-4
go get github.com/sahina/cvt/sdks/go@v0.3.0
go mod tidy
```

**Step 3: Verify `go.mod` contains the right dependency**

`consumer-4/go.mod` should look like:

```
module github.com/sahina/cvt-demo/consumer-4

go 1.25.0

require github.com/sahina/cvt/sdks/go v0.3.0

require (
    // indirect dependencies from the SDK...
)
```

**Step 4: Discover consumer-side Go SDK packages**

```bash
cd consumer-4
# List available packages in the SDK
go doc github.com/sahina/cvt/sdks/go/cvt
# Look for consumer-specific package:
go doc github.com/sahina/cvt/sdks/go/cvt/consumer 2>/dev/null || echo "no consumer package"
go doc github.com/sahina/cvt/sdks/go/cvt/consumer/adapters 2>/dev/null || echo "no adapters package"
```

Note the available types for mock client and HTTP adapter. The core `cvt.Validator` already provides `Validate`, `RegisterSchema`, `RegisterConsumer`, `CanIDeploy`, `ListConsumers` — these are confirmed from the producer tests.

**Step 5: Commit**

```bash
cd ..
git add consumer-4/go.mod consumer-4/go.sum
git commit -m "feat(consumer-4): scaffold Go module"
```

---

### Task 9: Consumer-4 — mock_test.go

**Files:**
- Create: `consumer-4/tests/mock_test.go`

**Step 1: Write `consumer-4/tests/mock_test.go`**

```go
// Package tests contains consumer-4 contract tests.
package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sahina/cvt/sdks/go/cvt"
	// VERIFY: import mock package — likely one of:
	//   github.com/sahina/cvt/sdks/go/cvt/consumer
	//   github.com/sahina/cvt/sdks/go/cvt/consumer/adapters
)

// TestMock_MultiplyResponse verifies mock generates valid response for /multiply
func TestMock_MultiplyResponse(t *testing.T) {
	validator, mock := setupMock(t)
	defer validator.Close()

	ctx := context.Background()

	// VERIFY: adjust mock.Fetch or mock.Get to match actual SDK API
	resp, err := mock.Fetch(ctx, "http://calculator-api/multiply?x=4&y=7")
	if err != nil {
		t.Fatalf("Mock fetch failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	body, _ := resp.Body()
	if _, ok := body["result"]; !ok {
		t.Errorf("Expected 'result' field in response, got: %v", body)
	}
}

// TestMock_DivideResponse verifies mock generates valid response for /divide
func TestMock_DivideResponse(t *testing.T) {
	validator, mock := setupMock(t)
	defer validator.Close()

	ctx := context.Background()

	resp, err := mock.Fetch(ctx, "http://calculator-api/divide?x=10&y=2")
	if err != nil {
		t.Fatalf("Mock fetch failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}

// TestMock_CapturesInteractions verifies interactions are recorded
func TestMock_CapturesInteractions(t *testing.T) {
	validator, mock := setupMock(t)
	defer validator.Close()

	ctx := context.Background()
	mock.ClearInteractions()

	mock.Fetch(ctx, "http://calculator-api/multiply?x=5&y=6")

	interactions := mock.GetInteractions()
	if len(interactions) != 1 {
		t.Errorf("Expected 1 interaction, got %d", len(interactions))
	}
}

// TestMock_CapturesAllConsumer4Endpoints verifies both endpoints are recorded
func TestMock_CapturesAllConsumer4Endpoints(t *testing.T) {
	validator, mock := setupMock(t)
	defer validator.Close()

	ctx := context.Background()
	mock.ClearInteractions()

	mock.Fetch(ctx, "http://calculator-api/multiply?x=4&y=7")
	mock.Fetch(ctx, "http://calculator-api/divide?x=10&y=2")

	interactions := mock.GetInteractions()
	if len(interactions) != 2 {
		t.Errorf("Expected 2 interactions, got %d", len(interactions))
	}
}

// TestMock_ResponseValidatesAgainstSchema verifies mock response matches schema
func TestMock_ResponseValidatesAgainstSchema(t *testing.T) {
	validator, mock := setupMock(t)
	defer validator.Close()

	ctx := context.Background()

	resp, err := mock.Fetch(ctx, "http://calculator-api/multiply?x=4&y=7")
	if err != nil {
		t.Fatalf("Mock fetch failed: %v", err)
	}

	body, _ := resp.Body()
	result, err := validator.Validate(ctx,
		cvt.ValidationRequest{Method: "GET", Path: "/multiply?x=4&y=7"},
		cvt.ValidationResponse{StatusCode: 200, Body: body},
	)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}
	if !result.Valid {
		t.Errorf("Mock response should validate against schema: %v", result.Errors)
	}
}

// setupMock creates a validator and mock client for testing.
// VERIFY: adjust return type to match actual mock client type.
func setupMock(t *testing.T) (*cvt.Validator, interface{ /* mock interface */ }) {
	t.Helper()
	config := getTestConfig(t)

	validator, err := cvt.NewValidator(config.CVTServerAddr)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	ctx := context.Background()
	if err := validator.RegisterSchema(ctx, "calculator-api", config.SchemaPath); err != nil {
		validator.Close()
		t.Fatalf("Failed to register schema: %v", err)
	}

	// VERIFY: adjust mock constructor — e.g.:
	//   consumer.NewMockClient(validator)
	//   adapters.NewMockClient(validator)
	mock := consumer.NewMockClient(validator)

	return validator, mock
}
```

> **IMPORTANT NOTE:** The Go consumer SDK mock/adapter API is not confirmed. Before writing this test, run `go doc github.com/sahina/cvt/sdks/go/cvt/consumer` to find the actual type and method names. Adjust all `// VERIFY:` comments accordingly.

**Step 2: Run test to see it compile (or find correct imports)**

```bash
cd consumer-4
CVT_SERVER_ADDR=localhost:9550 \
SCHEMA_PATH=../producer/calculator-api.json \
go test ./tests/... -run TestMock -v
```

Fix any import/compilation errors by adjusting package imports based on `go doc` output from Task 8.

**Step 3: Once passing, commit**

```bash
git add consumer-4/tests/mock_test.go
git commit -m "feat(consumer-4): add mock validation tests"
```

---

### Task 10: Consumer-4 — manual_test.go

**Files:**
- Create: `consumer-4/tests/manual_test.go`

**Step 1: Write `consumer-4/tests/manual_test.go`**

```go
package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sahina/cvt/sdks/go/cvt"
)

// TestManual_MultiplyValidation validates a real multiply response manually
func TestManual_MultiplyValidation(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()

	// Make real HTTP call
	url := fmt.Sprintf("%s/multiply?x=4&y=7", config.ProducerURL)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Manually validate
	result, err := validator.Validate(ctx,
		cvt.ValidationRequest{
			Method:  "GET",
			Path:    "/multiply?x=4&y=7",
			Headers: map[string]string{},
		},
		cvt.ValidationResponse{
			StatusCode: resp.StatusCode,
			Headers:    headersToMap(resp.Header),
			Body:       body,
		},
	)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid response: %v", result.Errors)
	}

	if val, ok := body["result"].(float64); !ok || val != 28 {
		t.Errorf("Expected result=28, got: %v", body["result"])
	}
}

// TestManual_DivideValidation validates a real divide response manually
func TestManual_DivideValidation(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()

	url := fmt.Sprintf("%s/divide?x=10&y=2", config.ProducerURL)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)

	result, err := validator.Validate(ctx,
		cvt.ValidationRequest{Method: "GET", Path: "/divide?x=10&y=2"},
		cvt.ValidationResponse{StatusCode: resp.StatusCode, Body: body},
	)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}
	if !result.Valid {
		t.Errorf("Expected valid: %v", result.Errors)
	}

	if val, ok := body["result"].(float64); !ok || val != 5 {
		t.Errorf("Expected result=5, got: %v", body["result"])
	}
}

// TestManual_DetectsInvalidResponse verifies schema violation is caught
func TestManual_DetectsInvalidResponse(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()

	result, err := validator.Validate(ctx,
		cvt.ValidationRequest{Method: "GET", Path: "/multiply?x=4&y=7"},
		cvt.ValidationResponse{
			StatusCode: 200,
			Body:       map[string]interface{}{"product": 28},
		},
	)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}
	if result.Valid {
		t.Error("Expected validation to fail for response with 'product' instead of 'result'")
	}
}

// TestManual_DivideByZeroErrorValidation validates a 400 error response
func TestManual_DivideByZeroErrorValidation(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()

	url := fmt.Sprintf("%s/divide?x=10&y=0", config.ProducerURL)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Fatalf("Expected 400 for division by zero, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)

	result, err := validator.Validate(ctx,
		cvt.ValidationRequest{Method: "GET", Path: "/divide?x=10&y=0"},
		cvt.ValidationResponse{StatusCode: resp.StatusCode, Body: body},
	)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}
	if !result.Valid {
		t.Errorf("400 error response should match error schema: %v", result.Errors)
	}
}
```

**Step 2: Create `consumer-4/tests/testutil_test.go`** (shared helpers)

```go
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
		// Resolve relative to this test file's directory
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
```

**Step 3: Run manual tests**

```bash
cd consumer-4
CVT_SERVER_ADDR=localhost:9550 \
PRODUCER_URL=http://localhost:10001 \
SCHEMA_PATH=../producer/calculator-api.json \
go test ./tests/... -run TestManual -v
```

Expected: All 4 `TestManual_*` tests PASS.

**Step 4: Commit**

```bash
git add consumer-4/tests/manual_test.go consumer-4/tests/testutil_test.go
git commit -m "feat(consumer-4): add manual validation tests and test utilities"
```

---

### Task 11: Consumer-4 — adapter_test.go

**Files:**
- Create: `consumer-4/tests/adapter_test.go`

**Step 1: Write `consumer-4/tests/adapter_test.go`**

```go
package tests

import (
	"context"
	"fmt"
	"testing"

	// VERIFY: import HTTP adapter package — likely:
	//   github.com/sahina/cvt/sdks/go/cvt/consumer/adapters
	//   github.com/sahina/cvt/sdks/go/cvt/consumer
)

// TestAdapter_MultiplyAutoValidation tests auto-validation for multiply
func TestAdapter_MultiplyAutoValidation(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()

	// VERIFY: construct the HTTP adapter — e.g.:
	//   adapters.NewValidatingClient(validator, true)
	//   consumer.NewValidatingHTTPClient(validator)
	client := adapters.NewValidatingClient(validator, true /* autoValidate */)
	defer client.ClearInteractions()

	url := fmt.Sprintf("%s/multiply?x=6&y=7", config.ProducerURL)
	resp, err := client.Get(ctx, url)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	interactions := client.GetInteractions()
	if len(interactions) == 0 {
		t.Fatal("Expected at least 1 interaction")
	}

	// VERIFY: adjust field access to match interaction struct
	last := interactions[len(interactions)-1]
	if !last.ValidationResult.Valid {
		t.Errorf("Expected valid interaction: %v", last.ValidationResult.Errors)
	}
}

// TestAdapter_DivideAutoValidation tests auto-validation for divide
func TestAdapter_DivideAutoValidation(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()
	client := adapters.NewValidatingClient(validator, true)
	client.ClearInteractions()

	url := fmt.Sprintf("%s/divide?x=20&y=4", config.ProducerURL)
	resp, err := client.Get(ctx, url)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	interactions := client.GetInteractions()
	if len(interactions) != 1 {
		t.Errorf("Expected 1 interaction, got %d", len(interactions))
	}
}

// TestAdapter_CapturesMultipleInteractions tests interaction recording
func TestAdapter_CapturesMultipleInteractions(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()
	client := adapters.NewValidatingClient(validator, true)
	client.ClearInteractions()

	client.Get(ctx, fmt.Sprintf("%s/multiply?x=3&y=4", config.ProducerURL))
	client.Get(ctx, fmt.Sprintf("%s/divide?x=8&y=2", config.ProducerURL))

	interactions := client.GetInteractions()
	if len(interactions) != 2 {
		t.Errorf("Expected 2 interactions, got %d", len(interactions))
	}

	for _, i := range interactions {
		if !i.ValidationResult.Valid {
			t.Errorf("Interaction should be valid: %v", i.ValidationResult.Errors)
		}
	}
}
```

**Step 2: Run adapter tests**

```bash
cd consumer-4
CVT_SERVER_ADDR=localhost:9550 \
PRODUCER_URL=http://localhost:10001 \
SCHEMA_PATH=../producer/calculator-api.json \
go test ./tests/... -run TestAdapter -v
```

Expected: 3 `TestAdapter_*` tests PASS.

**Step 3: Commit**

```bash
git add consumer-4/tests/adapter_test.go
git commit -m "feat(consumer-4): add adapter validation tests"
```

---

### Task 12: Consumer-4 — registration_test.go

**Files:**
- Create: `consumer-4/tests/registration_test.go`

**Step 1: Write `consumer-4/tests/registration_test.go`**

```go
package tests

import (
	"context"
	"testing"

	"github.com/sahina/cvt/sdks/go/cvt"
)

// TestRegistration_CaptureInteractions verifies interactions can be captured for registration
func TestRegistration_CaptureInteractions(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()
	mock := newMockClient(t, validator)
	mock.ClearInteractions()

	mock.Fetch(ctx, "http://calculator-api/multiply?x=4&y=7")
	mock.Fetch(ctx, "http://calculator-api/divide?x=10&y=2")

	interactions := mock.GetInteractions()
	if len(interactions) != 2 {
		t.Fatalf("Expected 2 interactions, got %d", len(interactions))
	}

	// VERIFY: adjust RegisterConsumerOptions struct to match SDK
	opts, err := validator.BuildConsumerFromInteractions(ctx, interactions, cvt.AutoRegisterConfig{
		ConsumerID:      config.ConsumerID,
		ConsumerVersion: config.Version,
		Environment:     config.Environment,
		SchemaVersion:   "1.0.0",
		SchemaID:        "calculator-api",
	})
	if err != nil {
		t.Fatalf("BuildConsumerFromInteractions failed: %v", err)
	}

	if opts.ConsumerID != config.ConsumerID {
		t.Errorf("Expected consumer ID %s, got %s", config.ConsumerID, opts.ConsumerID)
	}
	if len(opts.UsedEndpoints) != 2 {
		t.Errorf("Expected 2 endpoints, got %d", len(opts.UsedEndpoints))
	}
}

// TestRegistration_RegisterConsumerFromInteractions tests full auto-registration
func TestRegistration_RegisterConsumerFromInteractions(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()
	mock := newMockClient(t, validator)
	mock.ClearInteractions()

	mock.Fetch(ctx, "http://calculator-api/multiply?x=1&y=2")
	mock.Fetch(ctx, "http://calculator-api/divide?x=4&y=2")

	interactions := mock.GetInteractions()

	consumer, err := validator.RegisterConsumerFromInteractions(ctx, interactions, cvt.AutoRegisterConfig{
		ConsumerID:      config.ConsumerID,
		ConsumerVersion: config.Version,
		Environment:     config.Environment,
		SchemaVersion:   "1.0.0",
		SchemaID:        "calculator-api",
	})
	if err != nil {
		t.Fatalf("RegisterConsumerFromInteractions failed: %v", err)
	}

	if consumer.ConsumerID != config.ConsumerID {
		t.Errorf("Expected consumer ID %s, got %s", config.ConsumerID, consumer.ConsumerID)
	}
}

// TestRegistration_ManualRegistration tests explicit endpoint registration
func TestRegistration_ManualRegistration(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()

	// VERIFY: adjust RegisterConsumerOptions struct to match SDK
	consumer, err := validator.RegisterConsumer(ctx, cvt.RegisterConsumerOptions{
		ConsumerID:      config.ConsumerID,
		ConsumerVersion: config.Version,
		SchemaID:        "calculator-api",
		SchemaVersion:   "1.0.0",
		Environment:     config.Environment,
		UsedEndpoints: []cvt.EndpointUsage{
			{Method: "GET", Path: "/multiply", UsedFields: []string{"result"}},
			{Method: "GET", Path: "/divide", UsedFields: []string{"result"}},
		},
	})
	if err != nil {
		t.Fatalf("RegisterConsumer failed: %v", err)
	}

	if consumer.ConsumerID != config.ConsumerID {
		t.Errorf("Expected consumer ID %s, got %s", config.ConsumerID, consumer.ConsumerID)
	}
}

// TestRegistration_ListConsumers tests listing registered consumers
func TestRegistration_ListConsumers(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()

	// Register first
	validator.RegisterConsumer(ctx, cvt.RegisterConsumerOptions{
		ConsumerID:      config.ConsumerID,
		ConsumerVersion: config.Version,
		SchemaID:        "calculator-api",
		SchemaVersion:   "1.0.0",
		Environment:     config.Environment,
		UsedEndpoints: []cvt.EndpointUsage{
			{Method: "GET", Path: "/multiply", UsedFields: []string{"result"}},
			{Method: "GET", Path: "/divide", UsedFields: []string{"result"}},
		},
	})

	consumers, err := validator.ListConsumers(ctx, "calculator-api", config.Environment)
	if err != nil {
		t.Fatalf("ListConsumers failed: %v", err)
	}

	found := false
	for _, c := range consumers {
		if c.ConsumerID == config.ConsumerID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Consumer %s not found in list", config.ConsumerID)
	}
}

// TestRegistration_CanIDeploy tests deployment safety check
func TestRegistration_CanIDeploy(t *testing.T) {
	config := getTestConfig(t)
	validator := newTestValidator(t, config)
	defer validator.Close()

	ctx := context.Background()

	validator.RegisterConsumer(ctx, cvt.RegisterConsumerOptions{
		ConsumerID:      config.ConsumerID,
		ConsumerVersion: config.Version,
		SchemaID:        "calculator-api",
		SchemaVersion:   "1.0.0",
		Environment:     config.Environment,
		UsedEndpoints: []cvt.EndpointUsage{
			{Method: "GET", Path: "/multiply", UsedFields: []string{"result"}},
			{Method: "GET", Path: "/divide", UsedFields: []string{"result"}},
		},
	})

	result, err := validator.CanIDeploy(ctx, "calculator-api", "1.0.0", config.Environment)
	if err != nil {
		t.Logf("CanIDeploy returned error (expected if no consumers): %v", err)
		return
	}

	t.Logf("CanIDeploy result: SafeToDeploy=%v, Summary=%s", result.SafeToDeploy, result.Summary)
}

// newMockClient creates a mock client for registration tests.
// VERIFY: adjust return type to match actual mock client type.
func newMockClient(t *testing.T, validator *cvt.Validator) interface{ /* mock interface */ } {
	t.Helper()
	// VERIFY: adjust constructor
	return consumer.NewMockClient(validator)
}
```

**Step 2: Run registration tests**

```bash
cd consumer-4
CVT_SERVER_ADDR=localhost:9550 \
SCHEMA_PATH=../producer/calculator-api.json \
go test ./tests/... -run TestRegistration -v
```

Expected: All `TestRegistration_*` tests PASS.

**Step 3: Commit**

```bash
git add consumer-4/tests/registration_test.go
git commit -m "feat(consumer-4): add consumer registration tests"
```

---

### Task 13: Consumer-4 — main.go CLI

**Files:**
- Create: `consumer-4/main.go`

**Step 1: Write `consumer-4/main.go`**

```go
// Package main implements a CLI consumer for add and subtract operations.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/sahina/cvt/sdks/go/cvt"
)

var (
	producerURL   = envOrDefault("PRODUCER_URL", "http://localhost:10001")
	cvtServerAddr = envOrDefault("CVT_SERVER_ADDR", "localhost:9550")
	schemaPath    = envOrDefault("SCHEMA_PATH", "./calculator-api.json")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: consumer4 <command> <x> <y> [--validate]\n")
		fmt.Fprintf(os.Stderr, "Commands: add, subtract\n")
	}

	if len(os.Args) < 4 {
		flag.Usage()
		os.Exit(1)
	}

	command := os.Args[1]
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

	switch command {
	case "add":
		doAdd(x, y, validate)
	case "subtract":
		doSubtract(x, y, validate)
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func doAdd(x, y float64, validate bool) {
	path := fmt.Sprintf("/add?x=%v&y=%v", x, y)
	body := callAPI(path, validate)
	result := parseResult(body)
	fmt.Printf("%.0f + %.0f = %.0f\n", x, y, result)
}

func doSubtract(x, y float64, validate bool) {
	path := fmt.Sprintf("/subtract?x=%v&y=%v", x, y)
	body := callAPI(path, validate)
	result := parseResult(body)
	fmt.Printf("%.0f - %.0f = %.0f\n", x, y, result)
}

func callAPI(path string, validate bool) map[string]interface{} {
	resp, err := http.Get(producerURL + path)
	if err != nil {
		log.Fatalf("Error: No response from server. Is the producer running? %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}

	if resp.StatusCode >= 400 {
		if errMsg, ok := body["error"].(string); ok {
			fmt.Fprintln(os.Stderr, "Error:", errMsg)
		} else {
			fmt.Fprintf(os.Stderr, "Error: HTTP %d\n", resp.StatusCode)
		}
		os.Exit(1)
	}

	if validate {
		runValidation(path, resp.StatusCode, resp.Header, body)
	}

	return body
}

func runValidation(path string, statusCode int, headers http.Header, body map[string]interface{}) {
	validator, err := cvt.NewValidator(cvtServerAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to create CVT validator: %v\n", err)
		return
	}
	defer validator.Close()

	ctx := context.Background()
	if err := validator.RegisterSchema(ctx, "calculator-api", schemaPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to register schema: %v\n", err)
		return
	}

	hMap := make(map[string]string)
	for k, v := range headers {
		if len(v) > 0 {
			hMap[k] = v[0]
		}
	}

	result, err := validator.Validate(ctx,
		cvt.ValidationRequest{Method: "GET", Path: path, Headers: map[string]string{}},
		cvt.ValidationResponse{StatusCode: statusCode, Headers: hMap, Body: body},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: CVT validation error: %v\n", err)
		return
	}

	if !result.Valid {
		fmt.Fprintf(os.Stderr, "CVT Validation failed: %v\n", result.Errors)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "CVT validation enabled")
}

func parseResult(body map[string]interface{}) float64 {
	val, ok := body["result"].(float64)
	if !ok {
		log.Fatalf("Unexpected response format: missing 'result' field")
	}
	return val
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
```

**Step 2: Build and smoke-test**

```bash
cd consumer-4
go build -o consumer4 .

./consumer4 add 5 3
# Expected: 5 + 3 = 8

./consumer4 subtract 10 4
# Expected: 10 - 4 = 6

./consumer4 add 5 3 --validate
# Expected: CVT validation enabled\n5 + 3 = 8
```

**Step 3: Commit**

```bash
git add consumer-4/main.go
git commit -m "feat(consumer-4): add CLI main entry point"
```

---

### Task 14: Consumer-4 — Dockerfile

**Files:**
- Create: `consumer-4/Dockerfile`

**Step 1: Write `consumer-4/Dockerfile`**

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

WORKDIR /build

# Copy go module files and download dependencies
COPY consumer-4/go.mod consumer-4/go.sum ./
RUN go mod download

# Copy source code
COPY consumer-4/main.go .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /consumer4 .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Copy binary and schema
COPY --from=builder /consumer4 .
COPY producer/calculator-api.json ./calculator-api.json

# Set default environment variables
ENV PRODUCER_URL=http://producer:10001
ENV CVT_SERVER_ADDR=cvt:9550
ENV SCHEMA_PATH=/app/calculator-api.json

ENTRYPOINT ["./consumer4"]
```

**Step 2: Build the image**

```bash
cd /path/to/cvt-demo
docker build -f consumer-4/Dockerfile -t consumer-4-test .
```

Expected: `Successfully built <image-id>`.

**Step 3: Smoke-test**

```bash
docker run --rm --network cvt-demo-network \
  -e PRODUCER_URL=http://producer:10001 \
  -e CVT_SERVER_ADDR=cvt:9550 \
  -e SCHEMA_PATH=/app/calculator-api.json \
  consumer-4-test add 5 3
# Expected: 5 + 3 = 8
```

**Step 4: Commit**

```bash
git add consumer-4/Dockerfile
git commit -m "feat(consumer-4): add multi-stage Dockerfile"
```

---

## Phase 3: Infrastructure

### Task 15: Update docker-compose.yml

**Files:**
- Modify: `docker-compose.yml`

**Step 1: Add consumer-3 and consumer-4 services**

Add after the `consumer-2` service block (before the `networks:` section):

```yaml
  # Consumer-3 - Java CLI (run as one-off container)
  consumer-3:
    build:
      context: .
      dockerfile: consumer-3/Dockerfile
    environment:
      - PRODUCER_URL=http://producer:10001
      - CVT_SERVER_ADDR=cvt:9550
      - SCHEMA_PATH=/app/calculator-api.json
    depends_on:
      - producer
    networks:
      - cvt-demo-network
    profiles:
      - cli

  # Consumer-4 - Go CLI (run as one-off container)
  consumer-4:
    build:
      context: .
      dockerfile: consumer-4/Dockerfile
    environment:
      - PRODUCER_URL=http://producer:10001
      - CVT_SERVER_ADDR=cvt:9550
      - SCHEMA_PATH=/app/calculator-api.json
    depends_on:
      - producer
    networks:
      - cvt-demo-network
    profiles:
      - cli
```

**Step 2: Verify compose config is valid**

```bash
docker compose config --quiet
```

Expected: no output (success), or a clean config dump with no errors.

**Step 3: Test consumer-3 and consumer-4 via compose**

```bash
# With services running (make up)
docker compose run --rm consumer-3 multiply 4 7
# Expected: 4 * 7 = 28

docker compose run --rm consumer-4 add 5 3
# Expected: 5 + 3 = 8
```

**Step 4: Commit**

```bash
git add docker-compose.yml
git commit -m "feat: add consumer-3 and consumer-4 Docker services"
```

---

### Task 16: Update Makefile

**Files:**
- Modify: `Makefile`

**Step 1: Update `.PHONY` line**

Replace the existing `.PHONY` declaration with:

```makefile
.PHONY: help build up down logs clean test-all test test-contracts \
	producer-up producer-down \
	consumer-1-add consumer-1-subtract \
	consumer-1-add-validate consumer-1-subtract-validate \
	consumer-2-add consumer-2-multiply consumer-2-divide \
	consumer-2-add-validate consumer-2-multiply-validate consumer-2-divide-validate \
	consumer-3-multiply consumer-3-divide \
	consumer-3-multiply-validate consumer-3-divide-validate \
	consumer-4-add consumer-4-subtract \
	consumer-4-add-validate consumer-4-subtract-validate \
	shell-producer shell-cvt test-producer-http \
	test-consumer-1 test-consumer-1-mock test-consumer-1-live test-consumer-1-registration \
	test-consumer-2 test-consumer-2-mock test-consumer-2-live test-consumer-2-registration \
	test-consumer-3 test-consumer-3-mock test-consumer-3-live test-consumer-3-registration \
	test-consumer-4 test-consumer-4-mock test-consumer-4-live test-consumer-4-registration \
	test-unit test-live demo-breaking-change \
	test-producer test-producer-compliance test-producer-middleware \
	test-producer-registry test-producer-integration
```

**Step 2: Add consumer-3 operation targets** (after consumer-2 validate section)

```makefile
# =============================================================================
# Consumer-3 Operations (without validation)
# Usage: make consumer-3-multiply x=4 y=7
# =============================================================================

consumer-3-multiply:
	docker compose run --rm consumer-3 multiply $(x) $(y)

consumer-3-divide:
	docker compose run --rm consumer-3 divide $(x) $(y)

# =============================================================================
# Consumer-3 Operations (with CVT validation)
# Usage: make consumer-3-multiply-validate x=4 y=7
# =============================================================================

consumer-3-multiply-validate:
	docker compose run --rm consumer-3 multiply $(x) $(y) --validate

consumer-3-divide-validate:
	docker compose run --rm consumer-3 divide $(x) $(y) --validate
```

**Step 3: Add consumer-4 operation targets**

```makefile
# =============================================================================
# Consumer-4 Operations (without validation)
# Usage: make consumer-4-add x=5 y=3
# =============================================================================

consumer-4-add:
	docker compose run --rm consumer-4 add $(x) $(y)

consumer-4-subtract:
	docker compose run --rm consumer-4 subtract $(x) $(y)

# =============================================================================
# Consumer-4 Operations (with CVT validation)
# Usage: make consumer-4-add-validate x=5 y=3
# =============================================================================

consumer-4-add-validate:
	docker compose run --rm consumer-4 add $(x) $(y) --validate

consumer-4-subtract-validate:
	docker compose run --rm consumer-4 subtract $(x) $(y) --validate
```

**Step 4: Add consumer-3 and consumer-4 test targets** (after consumer-2 test targets)

```makefile
# Consumer-3 (Java) Tests
test-consumer-3-mock:
	@echo "Running Consumer-3 mock tests (no producer needed)..."
	cd consumer-3 && mvn test -Dtest="MockValidationTest" -q

test-consumer-3-live:
	@echo "Running Consumer-3 live tests (requires producer)..."
	cd consumer-3 && mvn test -Dtest="ManualValidationTest,AdapterValidationTest" -q

test-consumer-3-registration:
	@echo "Running Consumer-3 registration tests..."
	cd consumer-3 && mvn test -Dtest="RegistrationTest" -q

test-consumer-3:
	@echo "Running all Consumer-3 tests..."
	cd consumer-3 && mvn test

# Consumer-4 (Go) Tests
test-consumer-4-mock:
	@echo "Running Consumer-4 mock tests (no producer needed)..."
	cd consumer-4 && go test ./tests/... -run TestMock -v

test-consumer-4-live:
	@echo "Running Consumer-4 live tests (requires producer)..."
	cd consumer-4 && go test ./tests/... -run "TestManual|TestAdapter" -v

test-consumer-4-registration:
	@echo "Running Consumer-4 registration tests..."
	cd consumer-4 && go test ./tests/... -run TestRegistration -v

test-consumer-4:
	@echo "Running all Consumer-4 tests..."
	cd consumer-4 && go test ./tests/... -v
```

**Step 5: Update aggregate targets to include new consumers**

Update `test-unit`:

```makefile
test-unit:
	@echo "Running all mock/unit tests (no services needed except CVT server)..."
	@$(MAKE) -s test-consumer-1-mock
	@$(MAKE) -s test-consumer-2-mock
	@$(MAKE) -s test-consumer-3-mock
	@$(MAKE) -s test-consumer-4-mock
	@echo ""
	@echo "All mock tests completed!"
```

Update `test-live`:

```makefile
test-live:
	@echo "Running all live tests (requires producer)..."
	@$(MAKE) -s test-consumer-1-live
	@$(MAKE) -s test-consumer-2-live
	@$(MAKE) -s test-consumer-3-live
	@$(MAKE) -s test-consumer-4-live
	@echo ""
	@echo "All live tests completed!"
```

Update `test-all` (add consumer-3/4 operations):

```makefile
test-all:
	@echo "Running all consumer operations..."
	@echo ""
	@echo "=== Consumer-1 (Node.js) ==="
	@$(MAKE) -s consumer-1-add
	@$(MAKE) -s consumer-1-subtract
	@echo ""
	@echo "=== Consumer-2 (Python) ==="
	@$(MAKE) -s consumer-2-add
	@$(MAKE) -s consumer-2-multiply
	@$(MAKE) -s consumer-2-divide
	@echo ""
	@echo "=== Consumer-3 (Java) ==="
	@$(MAKE) -s consumer-3-multiply
	@$(MAKE) -s consumer-3-divide
	@echo ""
	@echo "=== Consumer-4 (Go) ==="
	@$(MAKE) -s consumer-4-add
	@$(MAKE) -s consumer-4-subtract
	@echo ""
	@echo "All operations completed!"
```

Update `test-contracts` (add consumer-3/4 with validation):

```makefile
test-contracts:
	@echo "Running all consumer operations with CVT validation..."
	@echo ""
	@echo "=== Consumer-1 (Node.js) with validation ==="
	@$(MAKE) -s consumer-1-add-validate
	@$(MAKE) -s consumer-1-subtract-validate
	@echo ""
	@echo "=== Consumer-2 (Python) with validation ==="
	@$(MAKE) -s consumer-2-add-validate
	@$(MAKE) -s consumer-2-multiply-validate
	@$(MAKE) -s consumer-2-divide-validate
	@echo ""
	@echo "=== Consumer-3 (Java) with validation ==="
	@$(MAKE) -s consumer-3-multiply-validate
	@$(MAKE) -s consumer-3-divide-validate
	@echo ""
	@echo "=== Consumer-4 (Go) with validation ==="
	@$(MAKE) -s consumer-4-add-validate
	@$(MAKE) -s consumer-4-subtract-validate
	@echo ""
	@echo "All contract validations passed!"
```

**Step 6: Update help text** (add consumer-3/4 sections in the help target)

After the Consumer-2 Operations echo block, add:

```makefile
	@echo "Consumer-3 Operations (Java - multiply, divide):"
	@echo "  make consumer-3-multiply        - Run: multiply (default: 5 * 3)"
	@echo "  make consumer-3-divide          - Run: divide (default: 5 / 3)"
	@echo "  make consumer-3-multiply-validate - With CVT validation"
	@echo "  make consumer-3-divide-validate - With CVT validation"
	@echo ""
	@echo "Consumer-4 Operations (Go - add, subtract):"
	@echo "  make consumer-4-add             - Run: add (default: 5 + 3)"
	@echo "  make consumer-4-subtract        - Run: subtract (default: 5 - 3)"
	@echo "  make consumer-4-add-validate    - With CVT validation"
	@echo "  make consumer-4-subtract-validate - With CVT validation"
	@echo ""
```

After Consumer-2 contract tests in help, add:

```makefile
	@echo "  make test-consumer-3           - Run all Consumer-3 tests"
	@echo "  make test-consumer-3-mock      - Run Consumer-3 mock tests (no producer needed)"
	@echo "  make test-consumer-3-live      - Run Consumer-3 live tests (requires producer)"
	@echo "  make test-consumer-3-registration - Run Consumer-3 registration tests"
	@echo "  make test-consumer-4           - Run all Consumer-4 tests"
	@echo "  make test-consumer-4-mock      - Run Consumer-4 mock tests (no producer needed)"
	@echo "  make test-consumer-4-live      - Run Consumer-4 live tests (requires producer)"
	@echo "  make test-consumer-4-registration - Run Consumer-4 registration tests"
```

Also update `demo-breaking-change` to include consumer-3 and consumer-4:

```makefile
demo-breaking-change:
	...
	@echo "Running consumer-3 registration tests..."
	-cd consumer-3 && CVT_ENVIRONMENT=demo mvn test -Dtest="RegistrationTest" -q 2>/dev/null || true
	@echo ""
	@echo "Running consumer-4 registration tests..."
	-cd consumer-4 && CVT_ENVIRONMENT=demo go test ./tests/... -run TestRegistration 2>/dev/null || true
	...
	@echo "Expected result: UNSAFE - all four consumers will break"
	@echo "  - consumer-1 uses 'result' field in /add and /subtract"
	@echo "  - consumer-2 uses 'result' field in /add, /multiply, and /divide"
	@echo "  - consumer-3 uses 'result' field in /multiply and /divide"
	@echo "  - consumer-4 uses 'result' field in /add and /subtract"
```

**Step 7: Verify make targets work**

```bash
make help | grep consumer-3
make help | grep consumer-4
make test-consumer-3-mock   # (requires CVT server)
make test-consumer-4-mock
```

**Step 8: Commit**

```bash
git add Makefile
git commit -m "feat: add consumer-3 and consumer-4 Makefile targets"
```

---

## Phase 4: CI/CD

### Task 17: Update test.yml

**Files:**
- Modify: `.github/workflows/test.yml`

**Step 1: Add `consumer-3-tests` job** (after `consumer-2-tests`)

```yaml
  consumer-3-tests:
    name: Consumer-3 Tests (Java)
    runs-on: ubuntu-latest
    needs: producer-tests
    timeout-minutes: 30

    services:
      cvt-server:
        image: ghcr.io/${{ github.repository_owner }}/cvt:0.3.0
        ports:
          - 9550:9550
        options: >-
          --health-cmd "/bin/grpc_health_probe -addr=:9550"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout cvt-demo
        uses: actions/checkout@v4

      - name: Setup CVT
        uses: ./.github/actions/setup-cvt

      - name: Setup Java
        uses: actions/setup-java@v4
        with:
          java-version: '21'
          distribution: 'temurin'

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25.x'
          cache: true
          cache-dependency-path: producer/go.sum

      - name: Install Dependencies
        run: |
          cd consumer-3
          mvn dependency:go-offline -q

      - name: Wait for CVT Server
        run: cvt wait --server localhost:9550 --timeout 60

      - name: Run Mock Tests
        id: mock
        run: |
          set -o pipefail
          cd consumer-3
          mvn test -Dtest="MockValidationTest" 2>&1 | tee mock.log
          echo "output<<EOF" >> $GITHUB_OUTPUT
          cat mock.log >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT

      - name: Build Producer
        run: |
          cd producer
          go build -o calculator-api .

      - name: Start Producer
        run: |
          cd producer
          PORT=10001 \
          CVT_SERVER_ADDR=localhost:9550 \
          SCHEMA_PATH=./calculator-api.yaml \
          CVT_ENABLED=true \
          ./calculator-api &
          echo $! > /tmp/producer.pid
          echo "Producer started with PID $(cat /tmp/producer.pid)"

      - name: Wait for Producer
        run: |
          for i in {1..30}; do
            if curl -sf http://localhost:10001/health; then
              echo ""
              echo "Producer is ready"
              exit 0
            fi
            echo "Waiting for producer... ($i/30)"
            sleep 2
          done
          echo "Producer failed to start"
          exit 1

      - name: Run Live Tests
        id: live
        run: |
          set -o pipefail
          cd consumer-3
          mvn test -Dtest="ManualValidationTest,AdapterValidationTest,RegistrationTest" 2>&1 | tee live.log
          echo "output<<EOF" >> $GITHUB_OUTPUT
          cat live.log >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT

      - name: Stop Producer
        if: always()
        run: |
          if [ -f /tmp/producer.pid ]; then
            kill $(cat /tmp/producer.pid) 2>/dev/null || true
          fi

      - name: Generate Job Summary
        if: always()
        run: |
          cat >> $GITHUB_STEP_SUMMARY << 'EOF'
          ## Consumer-3 Tests (Java)

          <details>
          <summary>🧪 Mock Tests (CVT Server Only)</summary>

          ```
          ${{ steps.mock.outputs.output }}
          ```

          </details>

          <details>
          <summary>🔌 Live Tests (Adapter + Manual + Registration)</summary>

          ```
          ${{ steps.live.outputs.output }}
          ```

          </details>
          EOF
```

**Step 2: Add `consumer-4-tests` job** (after `consumer-3-tests`)

```yaml
  consumer-4-tests:
    name: Consumer-4 Tests (Go)
    runs-on: ubuntu-latest
    needs: producer-tests
    timeout-minutes: 30

    services:
      cvt-server:
        image: ghcr.io/${{ github.repository_owner }}/cvt:0.3.0
        ports:
          - 9550:9550
        options: >-
          --health-cmd "/bin/grpc_health_probe -addr=:9550"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout cvt-demo
        uses: actions/checkout@v4

      - name: Setup CVT
        uses: ./.github/actions/setup-cvt

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25.x'
          cache: true
          cache-dependency-path: consumer-4/go.sum

      - name: Install Dependencies
        run: |
          cd consumer-4
          go mod download

      - name: Wait for CVT Server
        run: cvt wait --server localhost:9550 --timeout 60

      - name: Run Mock Tests
        id: mock
        run: |
          set -o pipefail
          cd consumer-4
          go test ./tests/... -run TestMock -v 2>&1 | tee mock.log
          echo "output<<EOF" >> $GITHUB_OUTPUT
          cat mock.log >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT

      - name: Build Producer
        run: |
          cd producer
          go build -o calculator-api .

      - name: Start Producer
        run: |
          cd producer
          PORT=10001 \
          CVT_SERVER_ADDR=localhost:9550 \
          SCHEMA_PATH=./calculator-api.yaml \
          CVT_ENABLED=true \
          ./calculator-api &
          echo $! > /tmp/producer.pid
          echo "Producer started with PID $(cat /tmp/producer.pid)"

      - name: Wait for Producer
        run: |
          for i in {1..30}; do
            if curl -sf http://localhost:10001/health; then
              echo ""
              echo "Producer is ready"
              exit 0
            fi
            echo "Waiting for producer... ($i/30)"
            sleep 2
          done
          echo "Producer failed to start"
          exit 1

      - name: Run Live Tests
        id: live
        run: |
          set -o pipefail
          cd consumer-4
          go test ./tests/... -run "TestManual|TestAdapter|TestRegistration" -v 2>&1 | tee live.log
          echo "output<<EOF" >> $GITHUB_OUTPUT
          cat live.log >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT

      - name: Stop Producer
        if: always()
        run: |
          if [ -f /tmp/producer.pid ]; then
            kill $(cat /tmp/producer.pid) 2>/dev/null || true
          fi

      - name: Generate Job Summary
        if: always()
        run: |
          cat >> $GITHUB_STEP_SUMMARY << 'EOF'
          ## Consumer-4 Tests (Go)

          <details>
          <summary>🧪 Mock Tests (CVT Server Only)</summary>

          ```
          ${{ steps.mock.outputs.output }}
          ```

          </details>

          <details>
          <summary>🔌 Live Tests (Adapter + Manual + Registration)</summary>

          ```
          ${{ steps.live.outputs.output }}
          ```

          </details>
          EOF
```

**Step 3: Commit**

```bash
git add .github/workflows/test.yml
git commit -m "ci: add consumer-3 and consumer-4 test jobs to test workflow"
```

---

### Task 18: Update consumer-only-test.yml

**Files:**
- Modify: `.github/workflows/consumer-only-test.yml`

**Step 1: Add `consumer-3-tests` job** (after `consumer-2-tests`)

```yaml
  consumer-3-tests:
    name: Consumer-3 Mock Tests (Java)
    runs-on: ubuntu-latest
    timeout-minutes: 30

    services:
      cvt-server:
        image: ghcr.io/${{ github.repository_owner }}/cvt:0.3.0
        ports:
          - 9550:9550
        options: >-
          --health-cmd "/bin/grpc_health_probe -addr=:9550"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout cvt-demo
        uses: actions/checkout@v4

      - name: Setup CVT
        uses: ./.github/actions/setup-cvt

      - name: Setup Java
        uses: actions/setup-java@v4
        with:
          java-version: '21'
          distribution: 'temurin'

      - name: Install Dependencies
        run: |
          cd consumer-3
          mvn dependency:go-offline -q

      - name: Wait for CVT Server
        run: cvt wait --server localhost:9550 --timeout 60

      - name: Run Mock Tests
        id: mock
        run: |
          set -o pipefail
          cd consumer-3
          mvn test -Dtest="MockValidationTest" 2>&1 | tee mock.log
          echo "output<<EOF" >> $GITHUB_OUTPUT
          cat mock.log >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT

      - name: Generate Job Summary
        if: always()
        run: |
          cat >> $GITHUB_STEP_SUMMARY << 'EOF'
          ## Consumer-3 Mock Tests (Java)

          <details>
          <summary>🧪 Mock Tests (CVT Server Only)</summary>

          ```
          ${{ steps.mock.outputs.output }}
          ```

          </details>
          EOF
```

**Step 2: Add `consumer-4-tests` job**

```yaml
  consumer-4-tests:
    name: Consumer-4 Mock Tests (Go)
    runs-on: ubuntu-latest
    timeout-minutes: 30

    services:
      cvt-server:
        image: ghcr.io/${{ github.repository_owner }}/cvt:0.3.0
        ports:
          - 9550:9550
        options: >-
          --health-cmd "/bin/grpc_health_probe -addr=:9550"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout cvt-demo
        uses: actions/checkout@v4

      - name: Setup CVT
        uses: ./.github/actions/setup-cvt

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25.x'
          cache: true
          cache-dependency-path: consumer-4/go.sum

      - name: Install Dependencies
        run: |
          cd consumer-4
          go mod download

      - name: Wait for CVT Server
        run: cvt wait --server localhost:9550 --timeout 60

      - name: Run Mock Tests
        id: mock
        run: |
          set -o pipefail
          cd consumer-4
          go test ./tests/... -run TestMock -v 2>&1 | tee mock.log
          echo "output<<EOF" >> $GITHUB_OUTPUT
          cat mock.log >> $GITHUB_OUTPUT
          echo "EOF" >> $GITHUB_OUTPUT

      - name: Generate Job Summary
        if: always()
        run: |
          cat >> $GITHUB_STEP_SUMMARY << 'EOF'
          ## Consumer-4 Mock Tests (Go)

          <details>
          <summary>🧪 Mock Tests (CVT Server Only)</summary>

          ```
          ${{ steps.mock.outputs.output }}
          ```

          </details>
          EOF
```

**Step 3: Commit**

```bash
git add .github/workflows/consumer-only-test.yml
git commit -m "ci: add consumer-3 and consumer-4 mock-only jobs to consumer-only workflow"
```

---

## Phase 5: Documentation

### Task 19: Update README.md

**Files:**
- Modify: `README.md`

**Step 1: Update architecture diagram** (replace the 3-service diagram with 5-service)

```
                    ┌─────────────────┐
                    │   CVT Server    │
                    │   (port 9550)   │
                    └────────┬────────┘
                             │ gRPC
         ┌───────────────────┼───────────────────────────────┐
         │                   │                   │           │           │
         ▼                   ▼                   ▼           ▼           ▼
┌─────────────────┐ ┌─────────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│   Consumer-1    │ │    Producer     │ │  Consumer-2  │ │  Consumer-3  │ │  Consumer-4  │
│   (Node.js)     │ │   (Go + CVT)    │ │  (Python/uv) │ │   (Java 21)  │ │   (Go 1.25)  │
│   CLI Tool      │ │  port 10001     │ │   CLI Tool   │ │   CLI Tool   │ │   CLI Tool   │
│  add, subtract  │ │  4 endpoints    │ │ add,mult,div │ │ mult, divide │ │ add, subtract│
└─────────────────┘ └─────────────────┘ └──────────────┘ └──────────────┘ └──────────────┘
         │                   ▲                   │           │           │
         └───────────────────┴───────────────────┴───────────┴───────────┘
                                          HTTP calls
```

**Step 2: Add Consumer-3 and Consumer-4 component sections**

After the Consumer-2 section, add:

```markdown
### Consumer-3 (Java)

A CLI tool that calls the Calculator API for **multiply** and **divide** operations:

```bash
java -jar consumer3.jar multiply 4 7
java -jar consumer3.jar divide 10 2
```

### Consumer-4 (Go)

A CLI tool that calls the Calculator API for **add** and **subtract** operations:

```bash
./consumer4 add 5 3
./consumer4 subtract 10 4
```
```

**Step 3: Update Quick Start sections 4 and 5, add sections 6 and 7**

After "Run Consumer-2 (Python)" section add:

```markdown
### 6. Run Consumer-3 (Java)

```bash
# Without CVT validation (default: A=5, B=3)
make consumer-3-multiply
make consumer-3-divide

# With custom values
make consumer-3-multiply x=7 y=8
make consumer-3-divide x=100 y=4

# With CVT validation
make consumer-3-multiply-validate
make consumer-3-divide-validate x=20 y=4
```

### 7. Run Consumer-4 (Go)

```bash
# Without CVT validation (default: A=5, B=3)
make consumer-4-add
make consumer-4-subtract

# With custom values
make consumer-4-add x=10 y=20

# With CVT validation
make consumer-4-add-validate
make consumer-4-subtract-validate x=100 y=50
```
```

**Step 4: Update Consumer Contract Tests section**

Add consumer-3 and consumer-4 to the install/run commands, test files table, and Validation Approaches section.

**Step 5: Update Make Targets section**

Add consumer-3/4 entries to all relevant subsections (Consumer Operations, Consumer Contract Tests).

**Step 6: Update CI/CD Workflows section**

Update the workflow diagram from:
```
producer-tests → (consumer-1-tests || consumer-2-tests)
```
to:
```
producer-tests → (consumer-1-tests || consumer-2-tests || consumer-3-tests || consumer-4-tests)
```

**Step 7: Update Project Structure tree**

Add `consumer-3/` and `consumer-4/` to the file tree.

**Step 8: Update Local Development section**

Add Consumer-3 and Consumer-4 local dev instructions.

**Step 9: Update Breaking Change Demo section**

Update `make demo-breaking-change` expected output to mention all 4 consumers.

**Step 10: Commit**

```bash
git add README.md
git commit -m "docs: update README for consumer-3 (Java) and consumer-4 (Go)"
```

---

### Task 20: Update CLAUDE.md and Final Commit

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Update Common Commands section**

Add to Testing section:
```bash
# Consumer-specific tests
make test-consumer-3      # All Consumer-3 tests
make test-consumer-4      # All Consumer-4 tests
make test-consumer-3-mock # Mock tests only
make test-consumer-4-mock # Mock tests only
```

Add to Running Consumers section:
```bash
# Consumer-3 (Java) - multiply, divide
make consumer-3-multiply x=4 y=7
make consumer-3-divide-validate x=10 y=2

# Consumer-4 (Go) - add, subtract
make consumer-4-add x=5 y=3
make consumer-4-subtract-validate x=10 y=4
```

**Step 2: Update Key Files table**

Add:
```
| consumer-3/src/main/java/demo/consumer3/Main.java | Java CLI with optional CVT validation     |
| consumer-3/pom.xml                                | Maven build, Java 21, JUnit 5             |
| consumer-3/src/test/java/demo/consumer3/          | Consumer-3 test suites                    |
| consumer-4/main.go                                | Go CLI with optional CVT validation        |
| consumer-4/go.mod                                 | Go module, cvt/sdks/go@v0.3.0             |
| consumer-4/tests/                                 | Consumer-4 test suites                    |
```

**Step 3: Update SDK Dependencies section**

Already has Go, Node, Python — confirm the Java entry:
```
- **Java SDK**: `io.github.sahina:cvt-sdk:0.3.0` via Maven Central
```

**Step 4: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md for consumer-3 and consumer-4"
```

---

## Final Verification

**Step 1: Run all mock tests (no producer needed)**

```bash
make test-unit
# Expected: All mock tests pass for consumer-1, 2, 3, 4
```

**Step 2: Run all live tests (requires `make up`)**

```bash
make up
make test
# Expected: All tests pass
```

**Step 3: Verify operations work end-to-end**

```bash
make test-all      # All consumer operations
make test-contracts # All with validation
```

**Step 4: Verify help output**

```bash
make help | grep -E "consumer-[34]"
```

Expected: Consumer-3 and Consumer-4 targets all listed.

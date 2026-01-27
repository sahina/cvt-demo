# CVT Presentation - Speaker Notes

Private talking points for the presenter. Titles match `presentation.md`.

---

## Pre-Demo Checklist

- [ ] Run `make up && make down` beforehand to pull Docker images
- [ ] Have backup video ready (record successful demo run)
- [ ] Terminal font size increased for visibility
- [ ] VS Code with Mermaid extension open for presentation.md

**Two-Phase Demo Reminder:**

- **Phase 1 (Consumer Testing):** Run `make test-consumer-1-mock` - no services needed
- **Phase 2 (Breaking Change):** Run `make up` live, then register + can-i-deploy

---

## Section 1: Opening & Context (0-5 min)

### Title

**Opening:**

- "I want to show you something we've been working on — a way to catch breaking API changes before they hit production."

**Set expectations:**

- Real code, real tests, live demo
- Keep it interactive — stop me if you have questions

---

### Why CVT?

**Tell the story (The Problem):**

- "I want to highlight few real world scenarios in API development."
- "Imagine you maintain a Calculator API. Running for a while, used by three teams."
- "You decide to rename 'result' to 'value' - more descriptive. Ship it Friday..."
- _Pause_ "Monday morning: Slack messages, incidents, unhappy VPs."

**Walk through problem bullets:**

- Producers don't know who consumes their APIs
- Breaking changes are missing in tests and discovered in production
- No single source of truth

**Transition to solution:**

- "CVT attempts to solves these problems"

**Walk through solution bullets:**

- Consumer Registry tracks dependencies
- can-i-deploy blocks breaking changes before merge
- OpenAPI specs as the source of truth

---

### Where CVT Fits - The Testing Pyramid

**Clarify positioning:**

- "CVT is NOT replacing E2E tests"

**Explain layers:**

- Unit tests: fast, isolated, function logic
- Integration: component interaction
- E2E: slow, expensive, business flows
- Contract tests: fast like unit, but validate API boundaries

**Value prop:**

- "Catch breaking changes early, before expensive E2E, before deployment"

---

## Section 2: Architecture (7-12 min)

### CVT Architecture

**Walk through components:**

1. **CVT Server:** "The brain - stores schemas, knows who consumes what"
2. **Producer:** "Calculator API in Go with CVT middleware"
3. **Consumers:** "Notice: polyglot! Node.js AND Python work seamlessly"

**Key insight:**

- "CVT acts as a contract registry - consumers register what they use, producers register schemas"

---

## Section 3: Consumer Testing (12-30 min)

### Three Consumer Validation Approaches

**Quick overview:**

- **Manual:** "Full control. More code, but flexible."
- **HTTP Adapter:** "Wraps your HTTP client. Zero test code changes."
- **Mock Client:** "THIS is the game-changer. No HTTP at all. CVT generates responses from schema."

---

## Deep Dive: Consumer Validation Approaches (Q&A Reference)

**Who uses consumer validation?**

- Teams that consume external or internal APIs
- Frontend teams calling backend services
- Microservice teams depending on other services
- Anyone who writes code like `fetch('/api/users')` or `axios.get('/orders')`

**Why consumer validation matters:**

- Catch contract violations in YOUR code before deployment
- Ensure your assumptions about the API match reality
- Get fast feedback without waiting for integration environments
- Document exactly which endpoints and fields you depend on

---

### Manual Validation — Deep Dive

**What it is:**
You explicitly call `validator.validate(request, response)` after making an HTTP call. You construct the validation objects yourself.

**Who should use it:**

- Teams that need to test specific error scenarios (400s, 500s, edge cases)
- When you want to validate only certain calls, not everything
- When testing negative cases (invalid inputs, expected failures)

**How it works:**

1. Make HTTP request normally
2. Capture the request details (method, path, headers, body)
3. Capture the response (status, headers, body)
4. Call `validator.validate(request, response)`
5. Assert on the validation result

**Pros:**

- Full control over what gets validated
- Can test error responses and edge cases
- Can skip validation for certain calls

**Cons:**

- Most verbose — more boilerplate code
- Easy to forget to validate some calls
- Test code is tightly coupled to validation logic

**Common questions:**

_"When would I choose manual over adapter?"_
When you need to test specific error responses (like 404s, 422s) or when you only want to validate certain calls, not all of them.

_"Can I mix manual with adapter?"_
Yes. Use adapter for happy-path tests, manual for edge cases.

_"Is manual validation slower than adapter?"_
No, they're about the same speed. The difference is in code verbosity, not performance.

_"How do I validate error responses like 400 or 500?"_
Manual is ideal for this. You make the call expecting an error, capture the response, and validate it against the error schema defined in your OpenAPI spec.

_"What if my OpenAPI spec doesn't define error responses?"_
Then CVT can't validate them. This is a good reason to define error schemas in your spec — it enables validation of unhappy paths too.

_"Can I use manual validation in production code, not just tests?"_
Technically yes, but it's not recommended. Validation adds overhead. Use it in tests; use middleware on the producer side for runtime validation.

---

### HTTP Adapter — Deep Dive

**What it is:**
A wrapper around your HTTP client (Axios, fetch, requests) that automatically validates every request/response against the contract.

**Who should use it:**

- Teams with existing test suites who want to add contract validation
- When you want "set it and forget it" validation
- Integration tests where you're testing real flows

**How it works:**

1. Wrap your HTTP client with the CVT adapter
2. Set `autoValidate: true`
3. Make HTTP calls as normal — your code doesn't change
4. Adapter intercepts responses and validates automatically
5. Access validation results via `adapter.getInteractions()`

**Pros:**

- Minimal code changes to existing tests
- Validates ALL calls automatically — nothing slips through
- Test code stays clean and focused on business logic

**Cons:**

- Less control — validates everything, even calls you might not care about
- Requires the real producer to be running
- Slower than mock — real network calls

**Common questions:**

_"What happens if validation fails?"_
By default, the adapter records the failure but doesn't throw. You check `getInteractions()` at the end of your test. You can configure it to throw immediately.

_"Does this slow down my tests?"_
Slightly — there's overhead for validation. But the network call itself is usually the bottleneck, not the validation.

_"Can I exclude certain endpoints from validation?"_
Yes, most adapters support include/exclude patterns.

_"What HTTP clients are supported?"_
Node.js: Axios, fetch, got. Python: requests, httpx. Go: net/http. Check the SDK docs for your language.

_"Can I use this with GraphQL?"_
CVT is designed for REST/OpenAPI. GraphQL has its own schema validation. You could potentially validate the HTTP transport layer, but not the GraphQL query semantics.

_"What if the producer is flaky or slow?"_
That's a real problem with adapter tests — they depend on the producer being available and healthy. Consider using mock tests for CI and adapter tests for integration environments where the producer is stable.

_"How do I see which validations failed?"_
Call `adapter.getInteractions()` after your tests. Each interaction has a `validationResult` with `valid: true/false` and `errors` array with details.

_"Can I use the adapter in a beforeAll/afterAll pattern?"_
Yes. Create the adapter in beforeAll, clear interactions in beforeEach, and assert all interactions valid in afterAll or afterEach.

---

### Mock Client — Deep Dive

**What it is:**
A fake HTTP client that doesn't make real network calls. Instead, CVT generates responses based on the OpenAPI schema.

**Who should use it:**

- CI/CD pipelines where you don't want to spin up dependencies
- Unit tests that need to be fast and isolated
- Testing before the producer API even exists (contract-first development)
- Teams that want deterministic, reproducible tests

**How it works:**

1. Create a mock adapter with the validator
2. Call `mock.fetch('/users/123')` — looks like a real HTTP call
3. CVT reads the OpenAPI schema for that endpoint
4. Generates a valid response (using examples if available, or fake data)
5. Returns the response — no network involved

**Pros:**

- Extremely fast — no network calls
- No producer needed — perfect for CI
- Deterministic — same input = same output
- Works before the API exists (contract-first)

**Cons:**

- You're testing against the SCHEMA, not the real API
- Won't catch bugs in the actual producer implementation
- Generated data may not be realistic (unless you use examples)
- Can give false confidence if schema doesn't match reality

**Common questions:**

_"If I'm not hitting the real API, what am I actually testing?"_
You're testing that YOUR code handles responses correctly according to the contract. You're not testing the producer — that's their responsibility.

_"What if the schema is wrong?"_
Then your mock tests will pass but live tests will fail. That's why you need BOTH mock and live tests. Mock for fast CI feedback, live for confidence.

_"Can I customize the generated responses?"_
Yes. Use `useExamples: true` to use examples from the schema, or provide custom fixtures.

_"How is this different from tools like MSW or Nock?"_
Those tools let you manually define mock responses. CVT Mock generates responses FROM the schema automatically. You don't maintain mock definitions — they come from the contract.

_"What if I need specific data in the response, not random generated data?"_
Use the `useExamples: true` option to pull from examples in your OpenAPI spec. Or provide custom fixtures for specific endpoints.

_"Can mock tests give false positives?"_
Yes. If the schema says the API returns `{result: number}` but the real API returns `{value: number}`, mock tests pass but live tests fail. Always pair mock with some live tests.

_"How do mock tests handle authentication?"_
They typically bypass it since there's no real server. If your code requires auth headers, you still need to provide them — the mock validates that your REQUEST is correct too.

_"Can I test timeout handling with mocks?"_
Not really — mocks return instantly. For timeout testing, use a real server or a specialized tool that can simulate delays.

_"What about testing retry logic?"_
Mocks return successfully every time by default. Some mock frameworks let you configure failures for specific calls to test retry behavior.

_"Should I use mock or live tests for CI?"_
Mock for fast feedback on every PR. Live tests either in CI (if you have stable test environments) or as a separate integration test suite that runs less frequently.

_"What's the recommended ratio of mock to live tests?"_
There's no magic number, but a common pattern: lots of mock tests (fast, cover many scenarios), fewer live tests (slower, cover critical paths and integration points).

---

### Manual Validation

**Walk through code:**

1. "Create validator, register schema"
2. "Make normal HTTP request"
3. "Build validation objects"
4. "Call validator.validate() explicitly"
5. "Assert result is valid"

**When to use:**

- Edge cases, error responses, precise control needed

---

### HTTP Adapter

**Key highlight:**

- "See autoValidate: true? That's it. Every request now automatically validated."

**Demo point:**

- "Your existing test code doesn't change. Just wrap the client."

**When to use:**

- Existing test suites, want contract validation with minimal changes

---

### Mock Client - The CI/CD Game-Changer

**The "aha" moment:**

- "No producer URL here. mock.fetch('<http://calculator-api/add>') - not a real URL!"
- "CVT reads schema, generates valid response. No network call, no producer."

---

### Demo: Mock vs Live — The Difference

**DEMO PHASE 1 - Show the contrast:**

```bash
# Setup: Start everything, then stop producer to show the difference
make up
make producer-down

# Step 1: Mock tests work — only need CVT server
make test-consumer-1-mock

# Step 2: Live tests fail — need producer
make test-consumer-1-live

# Step 3: Bring producer back
make producer-up

# Step 4: Now live tests work
make test-consumer-1-live
```

**Key points to make:**

- "I've stopped the producer but CVT server is still running."
- "Mock tests pass — CVT generates responses from the schema."
- "Now watch live tests..." (let it fail)
- "Right — live tests need the real API."
- (bring producer back) "Now they pass. Real HTTP calls, validated against the contract."

**The takeaway:**

- "Mock for CI — only needs CVT server, not your whole stack"
- "Live for full validation — real HTTP, catches issues mock can't"

---

### Python Consumer - Same Concepts

**Keep this brief:**

- "Python works the same way — same three patterns, just Pythonic syntax"
- Point to the table, don't read it
- Show the one-liner adapter example
- "If you're a Python shop, it just works"

**Don't linger here** — the audience has seen the pattern in Node.js

---

### The Glue: Consumer Registry + can-i-deploy

**This is the key slide — make sure the audience gets it:**

- "Before we jump to producer testing, let me show you how everything connects."
- "This is the magic that makes CVT more than just a validation tool."

**Walk through the diagram:**

1. "Consumer tests run — could be mock or live, doesn't matter"
2. "CVT records every interaction: this consumer called this endpoint, used these fields"
3. "That data goes to the Consumer Registry"
4. "Later, when the producer wants to deploy a new version..."
5. "CI runs can-i-deploy, which asks: will this change break anyone?"
6. "CVT checks the registry: 'Consumer A uses field X'... 'Oops, you're removing field X'... UNSAFE"

**Why this matters:**

- "Consumer validation ensures YOUR code is correct"
- "Producer validation ensures YOU don't break others"
- "The registry is the bridge between them"

**Analogy if needed:**

- "Think of it like a dependency graph, but for API contracts"
- "npm has package.json to track dependencies. CVT has the registry to track API consumers."

---

## Deep Dive: Consumer Registry & can-i-deploy (Q&A Reference)

### What Exactly Gets Recorded?

**Per interaction, CVT captures:**

- **Consumer identity** — name/id of the consuming application
- **Environment** — dev, staging, prod (configurable)
- **Timestamp** — when the interaction was recorded
- **HTTP method + path** — GET /users/{id}, POST /orders
- **Request details** — headers used, query params, path params, request body fields
- **Response details** — status code, response body fields accessed
- **Field-level usage** — which specific fields in the response the consumer actually uses

**Example recorded interaction:**

```json
{
  "consumer": "order-service",
  "environment": "prod",
  "schema": "user-api",
  "schemaVersion": "1.2.0",
  "endpoint": "GET /users/{id}",
  "requestFields": ["id"],
  "responseFields": ["id", "name", "email", "createdAt"],
  "statusCode": 200,
  "recordedAt": "2026-01-26T10:30:00Z"
}
```

### Why Is This Recorded?

**Primary purpose:** Enable can-i-deploy checks

- Without knowing what consumers use, you can't know what changes will break them

**Secondary benefits:**

- **API usage analytics** — which endpoints are popular? Which are unused?
- **Deprecation planning** — see who uses deprecated endpoints before removing
- **Documentation validation** — are consumers using the API as intended?
- **Dependency mapping** — build a graph of service dependencies

### How Detailed Is the Tracking?

**Field-level granularity:**

- Not just "Consumer A calls GET /users" but "Consumer A uses fields: id, name, email"
- This is critical — adding a field is safe, removing a USED field is breaking

**How CVT determines field usage:**

- For responses: tracks which fields appear in the response body
- More sophisticated implementations can track which fields the consumer code actually accesses (via SDK instrumentation)

**Depth of nested objects:**

- Tracks nested fields: `user.address.city`, `order.items[].productId`
- Array handling: knows if consumer uses array items, even if not every field in items

### Common Questions

_"What if the consumer uses fields that aren't in the schema?"_
CVT validates against the schema, so undocumented fields would fail validation. If validation passes, only schema-defined fields are being used.

_"Does CVT track request bodies too, or just responses?"_
Both. Request body fields, query parameters, headers — anything that's part of the contract can be tracked.

_"What happens if a consumer uses different fields in different test runs?"_
The registry accumulates interactions. If Test Run 1 uses fields A, B and Test Run 2 uses fields B, C — the registry knows the consumer uses A, B, C.

_"How does CVT identify the consumer?"_
Typically via configuration — you set `consumer: "order-service"` when initializing the SDK. Some setups derive it from service name or CI environment.

_"Can consumers lie about their identity?"_
Technically yes. It's a trust-based system. In practice, consumers have no incentive to lie — accurate registration protects THEM from breaking changes.

_"What about consumers that only exist in production?"_
If they don't run registration tests, you won't know about them. This is why registration should be part of CI for all consumers.

### Edge Cases & Gotchas

_"What if a consumer's tests don't cover all the fields they use in production?"_
You'll have blind spots. Encourage comprehensive tests, or use production traffic analysis as a supplement.

_"What if the same consumer has different usage patterns per environment?"_
Register per environment. Staging consumer might use different fields than prod consumer.

_"What about optional fields that consumers don't always use?"_
If a consumer CAN handle the field's absence, it's okay to remove it. But if their code crashes when it's missing, that's a break. CVT tracks what was observed, not what's tolerated.

_"How do we handle consumers that access the API directly (curl, Postman) without SDK?"_
Those won't be registered. CVT tracks programmatic consumers with the SDK. Ad-hoc usage isn't captured.

_"What if two teams have the same consumer name?"_
Namespace them: `team-a/order-service` vs `team-b/order-service`. Consumer identity should be unique.

_"Can I manually register a consumer without running tests?"_
Some CVT setups allow manual registration via API. Useful for onboarding legacy consumers.

_"What about versioned consumers? v1 uses different fields than v2."_
Register them as separate consumers: `order-service-v1`, `order-service-v2`. Or use metadata to distinguish.

### How can-i-deploy Uses the Registry

**The algorithm:**

1. Load the new schema version
2. Diff against the current/previous schema — find what changed
3. Categorize changes: addition, removal, modification
4. For each removal/breaking modification:
   - Query registry: "Who uses this endpoint/field?"
   - If anyone does → mark as affected
5. Return result: SAFE (no affected consumers) or UNSAFE (list affected)

---

### How can-i-deploy Detects Breaking Changes (Detailed)

**Step-by-step breakdown:**

**Step 1: Schema Diff**
CVT compares the OLD schema (currently deployed) vs NEW schema (proposed).

```text
OLD (v1.0.0):                    NEW (v2.0.0):
GET /users/{id}                  GET /users/{id}
Response:                        Response:
  - id: integer                    - id: integer
  - name: string                   - name: string
  - email: string        ←         (removed)
  - createdAt: string              - createdAt: string
                                   - phone: string  ← (added)
```

**Diff result:**

- REMOVED: `GET /users/{id}` response field `email`
- ADDED: `GET /users/{id}` response field `phone`

**Step 2: Query Registry for Affected Consumers**

CVT asks: "Who uses `GET /users/{id}` response field `email`?"

Registry data:

```text
consumer: order-service
  - GET /users/{id} → uses fields: [id, name, email]

consumer: notification-service
  - GET /users/{id} → uses fields: [id, email]

consumer: analytics-service
  - GET /users/{id} → uses fields: [id, createdAt]
```

**Match result:**

- `order-service` — uses `email` ❌ AFFECTED
- `notification-service` — uses `email` ❌ AFFECTED
- `analytics-service` — does NOT use `email` ✅ SAFE

**Step 3: Generate Report**

```text
can-i-deploy result: UNSAFE

Breaking changes detected:
  - REMOVED: GET /users/{id} response.email

Affected consumers:
  - order-service (uses: email)
  - notification-service (uses: email)

Safe changes:
  - ADDED: GET /users/{id} response.phone

Unaffected consumers:
  - analytics-service
```

---

### Breaking Change Detection Rules

**Response body changes:**

| Change                       | Breaking? | Why                                          |
| ---------------------------- | --------- | -------------------------------------------- |
| Remove field                 | YES       | Consumers reading it will get undefined/null |
| Rename field                 | YES       | Same as remove old + add new                 |
| Change type (string→number)  | YES       | Consumer parsing will fail                   |
| Make non-nullable → nullable | MAYBE     | Consumers not handling null will break       |
| Make nullable → non-nullable | NO        | Consumers already handle null                |
| Add optional field           | NO        | Consumers ignore unknown fields              |
| Add required field           | NO        | It's in response, not request                |

**Request body changes:**

| Change                   | Breaking? | Why                                 |
| ------------------------ | --------- | ----------------------------------- |
| Add required field       | YES       | Existing requests don't include it  |
| Remove field             | NO        | Server ignores extra fields         |
| Make optional → required | YES       | Consumers not sending it will fail  |
| Make required → optional | NO        | Consumers still send it, works fine |
| Change type              | YES       | Request validation will fail        |

**Endpoint changes:**

| Change                   | Breaking? | Why                               |
| ------------------------ | --------- | --------------------------------- |
| Remove endpoint          | YES       | Consumers calling it get 404      |
| Change path              | YES       | Same as remove + add              |
| Add endpoint             | NO        | Doesn't affect existing consumers |
| Change method (GET→POST) | YES       | Consumers using GET will fail     |

**Query/path parameter changes:**

| Change                   | Breaking? | Why                                   |
| ------------------------ | --------- | ------------------------------------- |
| Add required param       | YES       | Existing calls don't include it       |
| Remove param             | MAYBE     | If server relied on it, logic changes |
| Make optional → required | YES       | Calls without it will fail            |

---

### Concrete Examples

**Example 1: Safe change**

```yaml
# OLD
/users/{id}:
  get:
    responses:
      200:
        schema:
          properties:
            id: integer
            name: string

# NEW - adding optional field
/users/{id}:
  get:
    responses:
      200:
        schema:
          properties:
            id: integer
            name: string
            avatar: string  # ← NEW optional field
```

**Result:** SAFE — no consumer uses `avatar` (it didn't exist)

---

**Example 2: Breaking change — field removal**

```yaml
# OLD
/orders:
  get:
    responses:
      200:
        schema:
          properties:
            orderId: string
            status: string
            legacyCode: string  # ← will be removed

# NEW
/orders:
  get:
    responses:
      200:
        schema:
          properties:
            orderId: string
            status: string
            # legacyCode removed!
```

Registry shows `fulfillment-service` uses `legacyCode`.
**Result:** UNSAFE — `fulfillment-service` will break

---

**Example 3: Breaking change — type modification**

```yaml
# OLD
/products/{id}:
  get:
    responses:
      200:
        schema:
          properties:
            price: string  # "19.99"

# NEW
/products/{id}:
  get:
    responses:
      200:
        schema:
          properties:
            price: number  # 19.99
```

Registry shows `cart-service` uses `price`.
**Result:** UNSAFE — `cart-service` expects string, will get number

---

**Example 4: Breaking change — new required request field**

```yaml
# OLD
/orders:
  post:
    requestBody:
      schema:
        required: [customerId, items]
        properties:
          customerId: string
          items: array

# NEW
/orders:
  post:
    requestBody:
      schema:
        required: [customerId, items, paymentMethod]  # ← NEW required
        properties:
          customerId: string
          items: array
          paymentMethod: string
```

Registry shows `checkout-service` calls POST /orders.
**Result:** UNSAFE — `checkout-service` doesn't send `paymentMethod`

---

### What can-i-deploy CANNOT Detect

**Semantic changes:**

- Field still exists, same type, but MEANING changed
- Example: `status: "active"` used to mean X, now means Y
- CVT only sees structure, not semantics

**Behavioral changes:**

- Endpoint returns different data for same input
- Example: GET /users used to return all users, now returns paginated
- Schema might be the same, but behavior broke

**Performance changes:**

- Response time increased from 100ms to 5s
- Schema unchanged, but consumers timeout

**Side effects:**

- POST /orders used to send email, now it doesn't
- No schema change, but consumer expectation broken

**Undocumented consumers:**

- If they never registered, CVT doesn't know about them

_"So can-i-deploy isn't perfect?"_
Correct. It catches STRUCTURAL breaking changes. Semantic, behavioral, and performance changes require other testing strategies (integration tests, monitoring, etc.).

---

**What counts as "breaking":**

- Field removed → consumers using it will break
- Field type changed incompatibly (string → number) → likely break
- Field made required when it was optional → consumers not sending it will break
- Endpoint removed → consumers calling it will break

**What's usually safe:**

- Adding new optional fields
- Adding new endpoints
- Making required fields optional
- Adding new enum values (usually)

_"What about adding required fields to request bodies?"_
Breaking! Existing consumers aren't sending the new field. Their requests will fail validation.

_"What if I rename a field (remove old, add new with same data)?"_
Breaking. CVT sees it as removal + addition. Consumers using the old name will break.

_"How do I do a safe rename?"_

1. Add new field (keep old)
2. Wait for consumers to migrate
3. Deprecate old field
4. After all consumers migrate, remove old field
5. can-i-deploy will be SAFE once no one uses the old field

_"Can I override can-i-deploy if I know it's safe?"_
Yes — it's advisory, not a hard block (unless you configure CI to fail). But document why you're overriding.

_"What if the registry data is stale?"_
Set TTLs on registrations, or require consumers to re-register periodically. Stale data = false sense of safety.

---

### Transition to Producer Testing

**Quick breath:**

- "So that's consumer testing — three approaches, works in Node.js and Python"
- "Now let's flip to the producer side"

---

## Section 4: Producer Testing (30-38 min)

### Four Producer Validation Approaches

**Quick overview:**

- Schema Compliance: Unit test handlers against schema
- Middleware Modes: Test strict/warn/shadow behavior
- Consumer Registry: can-i-deploy checks
- HTTP Integration: Full end-to-end

---

## Deep Dive: Producer Validation Approaches (Q&A Reference)

**Who uses producer validation?**

- Teams that BUILD and MAINTAIN APIs
- Backend engineers who own services
- Platform teams providing APIs to internal consumers
- Anyone responsible for "don't break the clients"

**Why producer validation matters:**

- Ensure your API implementation matches the contract
- Catch schema drift before it reaches production
- Know who depends on you and what they use
- Block breaking changes before merge, not after deploy

---

### Schema Compliance Testing — Deep Dive

**What it is:**
Unit testing your API handlers directly (no HTTP server) and validating the responses against the OpenAPI schema.

**Who should use it:**

- API developers during active development
- Teams practicing TDD with contracts
- Anyone who wants fast feedback on schema compliance

**How it works:**

1. Create a test request using `httptest.NewRequest` (Go) or equivalent
2. Call your handler function directly — no server, no network
3. Capture the response
4. Send to CVT for validation against the schema
5. Assert validation passes

**Pros:**

- Extremely fast — no HTTP server, no network
- Runs in milliseconds
- Catches schema drift immediately during development
- Easy to integrate into existing unit test suites

**Cons:**

- Doesn't test the full HTTP stack (routing, middleware, serialization)
- Handler must be callable in isolation
- May miss issues that only appear in real HTTP context

**Common questions:**

_"How is this different from just running my API and hitting it?"_
Speed. Schema compliance tests run in milliseconds because there's no HTTP overhead. You can run them on every save, every commit.

_"What if my handler depends on a database?"_
Mock it, same as any unit test. The goal here is to validate that given certain inputs, your handler produces schema-compliant outputs.

_"Should I do this AND HTTP integration tests?"_
Yes. Schema compliance for fast feedback during development. HTTP integration for confidence that the full stack works.

_"Can I use this with any language/framework?"_
Yes, as long as you can call your handler directly and capture the response. Go, Node.js, Python, Java — all work.

_"What about request validation, not just response?"_
You can validate that your handler correctly rejects invalid requests. Send a bad request, check that the handler returns the appropriate error response.

_"How do I handle handlers that return streams or files?"_
Schema compliance is best for JSON responses. For streams/files, you'd typically validate headers and content-type, not the body content.

_"What if my handler calls other services?"_
Mock those dependencies. You're testing YOUR handler's behavior given certain inputs, not the entire distributed system.

_"Is this the same as snapshot testing?"_
No. Snapshot testing checks that output matches a saved snapshot. Schema compliance checks that output conforms to a contract. Schemas are more flexible — they allow any valid response, not just one specific response.

---

### Middleware Modes — Deep Dive

**What it is:**
CVT middleware that sits in front of your API and validates requests/responses in real-time. Three modes control what happens on violations.

**Who should use it:**

- Teams deploying CVT to existing production services
- Anyone who wants runtime contract enforcement
- Teams doing gradual rollouts of contract validation

**The Three Modes:**

**Strict Mode:**

- Invalid request → 400 Bad Request (blocked)
- Invalid response → 500 Internal Server Error (blocked)
- Use when: You want hard enforcement, zero tolerance for violations
- Risk: May break clients if your schema is wrong

**Warn Mode:**

- Invalid request → Log warning, continue processing
- Invalid response → Log warning, return response anyway
- Use when: Rolling out CVT to an existing service, want visibility before enforcement
- Benefit: See violations without breaking anything

**Shadow Mode:**

- Invalid request → Record metric only, continue silently
- Invalid response → Record metric only, return response
- Use when: Testing in production, want data before making decisions
- Benefit: Zero risk, pure observation

**Rollout strategy:**

1. Start with Shadow in production — collect data on violations
2. Review metrics — are violations real bugs or schema issues?
3. Fix schema or code as needed
4. Move to Warn — log violations, alert on them
5. Finally, Strict — enforce the contract

**Common questions:**

_"What if I enable Strict and my schema is wrong?"_
You'll break your clients. That's why you start with Shadow, validate the schema is accurate, then gradually tighten.

_"Does middleware add latency?"_
Yes, but typically <5ms for validation. Shadow mode is fastest since it's async.

_"Can I use different modes for different endpoints?"_
Depends on the implementation. Some setups allow per-route configuration.

_"What's the difference between validating requests vs responses?"_
Request validation protects YOUR service from bad input. Response validation ensures YOU'RE sending correct output. Both are valuable.

_"Should I validate requests, responses, or both?"_
Start with responses — that's where contract drift usually shows up. Add request validation if you want to enforce input contracts strictly.

_"How do I monitor violations in Shadow mode?"_
CVT emits metrics (Prometheus-compatible). Set up dashboards and alerts on validation failure counts. Review before moving to stricter modes.

_"What happens to in-flight requests when I switch modes?"_
Mode changes typically apply to new requests. In-flight requests complete under the old mode.

_"Can I roll back from Strict to Warn quickly?"_
Yes, it's a configuration change. Have a runbook ready — if Strict causes issues, you can quickly switch to Warn.

_"Should I use Strict in development and Shadow in production?"_
Common pattern: Strict in dev/staging (fail fast, find issues early), Shadow or Warn in production (observe before enforcing).

_"What if a violation is actually a schema bug, not a code bug?"_
Fix the schema! The contract should reflect reality. If your API returns something the schema doesn't allow, update the schema (carefully, ensuring backward compatibility).

---

### Consumer Registry & can-i-deploy — Deep Dive

**What it is:**
CVT tracks which consumers use which endpoints and fields. Before deploying a schema change, `can-i-deploy` checks if any consumers would break.

**Who should use it:**

- Producer teams about to deploy schema changes
- CI/CD pipelines for automated safety checks
- Release managers coordinating breaking changes

**How it works:**

1. Consumers run their tests with CVT — interactions are recorded
2. CVT builds a registry: "Consumer A uses GET /users, fields: id, name, email"
3. Producer wants to deploy v2.0.0 with schema changes
4. CI runs `cvt can-i-deploy --schema my-api --version 2.0.0`
5. CVT compares new schema against consumer interactions
6. Returns SAFE (no conflicts) or UNSAFE (lists affected consumers)

**What counts as a breaking change:**

- Removing a field that consumers use
- Renaming a field (removing old, adding new)
- Changing a field type (string → number)
- Removing an endpoint
- Changing required/optional status in incompatible ways

**What's usually NOT breaking:**

- Adding new optional fields
- Adding new endpoints
- Adding new optional parameters

**Pros:**

- Know BEFORE merge if you'll break someone
- Identifies exactly which consumers are affected
- Enables coordination: "Talk to Team X before deploying"

**Cons:**

- Only as good as consumer registrations — if consumers don't register, you don't know about them
- Requires consumers to actively participate
- Initial setup overhead

**Common questions:**

_"What if a consumer doesn't register?"_
You won't know about them. CVT only knows what's registered. Encourage all consumers to run registration tests.

_"Can I still deploy if can-i-deploy says UNSAFE?"_
Yes, it's a safety check, not a hard block (unless you configure CI to fail). But you should coordinate with affected teams first.

_"How do consumers register?"_
By running their tests with CVT. The test interactions are recorded and sent to the registry.

_"What about external consumers outside my org?"_
They'd need to register too, or you need another way to track them. CVT works best for internal APIs where you control both sides.

_"How granular is the tracking? Endpoint level? Field level?"_
Field level. CVT knows that Consumer A uses the `email` field on `GET /users/{id}`. So if you remove `email`, it flags Consumer A.

_"What if I need to make a breaking change?"_
CVT tells you WHO is affected. Then you coordinate: notify teams, agree on migration timeline, deprecate old version, give consumers time to update, then remove.

_"How do I handle versioned APIs?"_
Register each version as a separate schema. Consumers register against the version they use. You can deprecate old versions while maintaining them.

_"What if consumers use fields I didn't intend to be part of the contract?"_
That's a conversation to have. Either add those fields to your contract officially, or tell consumers they're using undocumented fields at their own risk.

_"Can I see a dashboard of all consumers and what they use?"_
Yes, CVT provides APIs to query the registry. Build dashboards showing consumer dependencies per endpoint.

_"What happens if a consumer's tests are flaky and don't always register?"_
Registrations should be consistent. If tests are flaky, fix them. Incomplete registration means incomplete protection.

_"How long are registrations retained?"_
Configurable. Common pattern: retain until explicitly removed, or TTL-based (e.g., 30 days since last registration).

---

### HTTP Integration Testing — Deep Dive

**What it is:**
Full end-to-end tests that make real HTTP calls to your running API and validate responses against the contract.

**Who should use it:**

- Teams that want confidence in the complete stack
- Pre-deployment smoke tests
- Integration test suites

**How it works:**

1. Start your API server
2. Make real HTTP requests to it
3. Capture responses
4. Validate against CVT schema
5. Assert validation passes

**Pros:**

- Tests the REAL thing — routing, middleware, serialization, everything
- Highest confidence that what you ship actually works
- Catches issues that unit tests miss

**Cons:**

- Slower — requires running server, network calls
- More infrastructure — need the service up
- Flakier — network issues, timing, dependencies

**When to use:**

- After schema compliance tests pass
- In CI before deployment
- As smoke tests after deployment

**Common questions:**

_"If I have schema compliance tests, do I need this?"_
Yes. Schema compliance tests your handler in isolation. HTTP integration tests the full stack. They catch different issues.

_"How often should I run these?"_
Every CI build at minimum. Some teams also run them as post-deployment smoke tests.

_"What about dependencies (databases, other services)?"_
Either use real dependencies (slower, more realistic) or mocks (faster, less realistic). Depends on what you're testing.

_"How do I balance test speed vs coverage?"_
Pyramid approach: lots of fast schema compliance tests, fewer HTTP integration tests. Integration tests cover critical paths; compliance tests cover breadth.

_"Can I run HTTP integration tests in parallel?"_
Yes, if your test environment supports it. Be careful about shared state (database, etc.) causing flakiness.

_"What's the difference between this and E2E tests?"_
HTTP integration tests validate contract compliance. E2E tests validate user journeys. E2E might span multiple services; HTTP integration typically targets one API.

_"Should HTTP integration tests hit a real database?"_
For contract validation, it often doesn't matter — you're checking response shape, not data correctness. But realistic data can help catch serialization issues.

_"How do I handle authentication in integration tests?"_
Either use test credentials, mock the auth layer, or run in an environment where auth is relaxed for testing.

_"What if my API has rate limiting?"_
Either disable rate limiting for test environments, use higher limits, or design tests to stay under limits.

---

### Schema Compliance Testing

**Key insight:**

- "Testing handlers DIRECTLY. No HTTP server. Create test request, call handler, validate response."

**Value:**

- "Catches schema drift at unit test level - fast feedback, early in development"

---

### Middleware Modes

**Explain each:**

- **Strict:** "Production mode. Violations blocked - 400 or 500 error."
- **Warn:** "Rollout mode. Violations logged, request continues. Good for adding CVT to existing service."
- **Shadow:** "Canary mode. Violations recorded as metrics only. Test in production safely."

**Show config:**

- "One line change: ModeStrict becomes ModeWarn"

```bash
make test-producer-middleware
```

---

## Section 5: Breaking Change Detection (38-48 min)

### Demo: Setup & Register Consumers

**Note:** Services should already be running from the consumer demo. If not, run `make up` first.

**Terminal:**

```bash
make test-consumer-1-registration
make test-consumer-2-registration
```

**Explain while running:**

- "These tests run and CVT records what each consumer uses."
- "Consumer-1: /add, /subtract. Consumer-2: /add, /multiply, /divide."
- "Importantly — both consumers read the `result` field."

---

### The Breaking Change

**Point to the diagram:**

- "V1 returns `{result: 8}`. A developer wants to rename it to `{value: 8}`. Seems harmless..."
- "But look — Consumer-1 does `response.data.result`. Consumer-2 does `response.json()['result']`."
- "Both will break. undefined in Node.js, KeyError in Python."

**Point to the diff:**

- "Here's the actual change. Version 1.0.0 to 2.0.0. Just one line — `result` becomes `value`."

---

### Demo: can-i-deploy

**This is the payoff:**

- "Okay, moment of truth. Can we deploy version 2.0.0?"

**Terminal:**

```bash
make demo-breaking-change
```

**Read the result:**

- "UNSAFE. Both consumers will break."
- "CVT tells us exactly who is affected and what they use."

**Key takeaway:**

- "This check runs in CI. This PR gets blocked before merge. Before anyone's Friday gets ruined."

---

## Section 6: Gotchas & Best Practices (48-52 min)

### Three Important Gotchas

1. **Schema Versioning:** "Always semantic versioning. Never overwrite - register new versions."
2. **Test Isolation:** "Mock tests in fast CI loop. Integration tests separate."
3. **Registration Timing:** "Consumers must register BEFORE can-i-deploy."

---

### Which Approach Should I Use?

**Walk through with audience:**

- Need to run without producer? → Mock
- Have existing test suite? → Adapter
- Testing edge cases? → Manual

**Ask about their scenarios**

---

## Section 7: Call to Action (52-55 min)

### Resources + Quick Start

**Point to resources:**

- CVT docs for full details
- Demo repo for hands-on learning

---

### Quick Recap

**Keep it casual:**

- Don't read the bullets — just summarize verbally
- "So: validate against schemas, track who uses what, block breaking changes, and fast CI with mocks"
- "That's the core idea"

**Transition to Q&A:**

- "What questions do you have?"

---

### Next Steps

1. "Clone the demo, run make up, run make test"
2. "Pick ONE API in your organization to pilot"
3. "Start a conversation with your team"

---

## Section 8: Q&A (50-60 min)

### Q&A

**Prepared responses:**

**"How does this compare to Pact?"**

- Different philosophies. Pact is consumer-driven - consumers define contracts.
- CVT is schema-first - start with OpenAPI.
- Both valid, choose based on workflow.

**"What about performance?"**

- Strict mode has overhead
- Shadow mode nearly invisible - async metrics
- Start with Shadow in production if concerned

**"gRPC/GraphQL support?"**

- Currently focused on REST/OpenAPI
- Check docs for updates

**"What about necessary breaking changes?"**

- CVT tells you WHO is affected
- Then coordinate with those teams on migration timing
- It's about informed decisions, not blocking all changes

---

### Thank You

**Closing:**

- Thank audience for their time
- Encourage questions after session
- Point to links for follow-up

---

## Terminal Commands Reference

### Setup & Health

```bash
make up                              # Start all services
make down                            # Stop all services
curl http://localhost:10001/health   # Health check
```

### Consumer Tests

```bash
make test-consumer-1                 # All Node.js tests
make test-consumer-1-mock            # Mock tests only (fast)
make test-consumer-2                 # All Python tests
make test-consumer-2-mock            # Mock tests only (fast)
make test-unit                       # All mock tests (both consumers)
make test-integration                # All integration tests
```

### Producer Tests

```bash
make test-producer                   # All producer tests
make test-producer-compliance        # Schema compliance
make test-producer-middleware        # Middleware modes
make test-producer-registry          # Consumer registry
```

### Breaking Change Demo

```bash
make test-consumer-1-registration    # Register Node.js consumer
make test-consumer-2-registration    # Register Python consumer
make demo-breaking-change            # Check if v2 is safe
```

---

## Emergency Backup Plan

If Docker fails during demo:

1. Switch to pre-recorded video
2. Walk through code in editor
3. Show screenshots of expected output

**Always have ready:**

- Screen recording of full demo
- Screenshots of test output
- Screenshots of can-i-deploy output

---

## Timing Guide

| Section                   | Time      | Key Moment             |
| ------------------------- | --------- | ---------------------- |
| Opening & Context         | 0-5 min   | Friday deploy story    |
| Architecture              | 5-10 min  | (no demo)              |
| Consumer Testing          | 10-25 min | **DEMO: Mock tests**   |
| Producer Testing          | 25-32 min | Middleware modes       |
| Breaking Change Detection | 32-45 min | **DEMO: can-i-deploy** |
| Gotchas & Call to Action  | 45-50 min | Decision flowchart     |
| Q&A                       | 50-60 min | **10 min buffer**      |

**Two-Phase Demo:**

1. **Phase 1:** `make test-consumer-1-mock` - shows tests run without producer
2. **Phase 2:** `make up` → register consumers → `make demo-breaking-change`

**Note:** Compressed Python section saves ~3-4 min. Use that buffer for Q&A or demos running long.

"""
CONSUMER REGISTRATION
---------------------
Registers which API endpoints your service uses so breaking changes can be detected.

Prerequisites:
- CVT server running

Key concepts:
- Auto-registration: Capture interactions from mocks/adapters, then register
- Manual registration: Explicitly declare used endpoints
- can-i-deploy: Check if a schema change is safe for all consumers
"""

import pytest

from cvt_sdk.adapters import MockSession
from cvt_sdk.auto_register import AutoRegisterConfig, EndpointUsage


class TestConsumerRegistration:

    @pytest.fixture(scope="class")
    def mock_session(self, validator):
        session = MockSession(validator, cache=True)
        yield session
        session.clear_cache()

    def test_capture_interactions_for_auto_registration(
        self, mock_session, validator, consumer_id, consumer_version, environment
    ):
        mock_session.clear_interactions()

        mock_session.get("http://calculator-api/add", params={"a": 5, "b": 3})
        mock_session.get("http://calculator-api/multiply", params={"a": 4, "b": 7})
        mock_session.get("http://calculator-api/divide", params={"a": 10, "b": 2})

        interactions = mock_session.get_interactions()
        assert len(interactions) == 3

        opts = validator.build_consumer_from_interactions(
            interactions,
            AutoRegisterConfig(
                consumer_id=consumer_id,
                consumer_version=consumer_version,
                environment=environment,
                schema_version="1.0.0",
                schema_id="calculator-api",
            ),
        )

        assert opts.consumer_id == consumer_id
        assert opts.used_endpoints is not None
        assert len(opts.used_endpoints) == 3

        endpoints = [f"{e.method} {e.path}" for e in opts.used_endpoints]
        assert "GET /add" in endpoints
        assert "GET /multiply" in endpoints
        assert "GET /divide" in endpoints

    def test_register_consumer_from_interactions(
        self, mock_session, validator, consumer_id, consumer_version, environment
    ):
        mock_session.clear_interactions()

        mock_session.get("http://calculator-api/add", params={"a": 1, "b": 2})
        mock_session.get("http://calculator-api/multiply", params={"a": 3, "b": 4})
        mock_session.get("http://calculator-api/divide", params={"a": 8, "b": 2})

        interactions = mock_session.get_interactions()

        consumer = validator.register_consumer_from_interactions(
            interactions,
            AutoRegisterConfig(
                consumer_id=consumer_id,
                consumer_version=consumer_version,
                environment=environment,
                schema_version="1.0.0",
                schema_id="calculator-api",
            ),
        )

        assert consumer is not None
        assert consumer["consumer_id"] == consumer_id


class TestManualRegistration:

    def test_register_consumer_with_explicit_endpoints(
        self, validator, consumer_id, consumer_version, environment
    ):
        from cvt_sdk import RegisterConsumerOptions

        consumer = validator.register_consumer(
            RegisterConsumerOptions(
                consumer_id=consumer_id,
                consumer_version=consumer_version,
                schema_id="calculator-api",
                schema_version="1.0.0",
                environment=environment,
                used_endpoints=[
                    EndpointUsage(method="GET", path="/add", used_fields=["result"]),
                    EndpointUsage(method="GET", path="/multiply", used_fields=["result"]),
                    EndpointUsage(method="GET", path="/divide", used_fields=["result"]),
                ],
            )
        )

        assert consumer is not None
        assert consumer["consumer_id"] == consumer_id

    def test_list_registered_consumers(
        self, validator, consumer_id, consumer_version, environment
    ):
        from cvt_sdk import RegisterConsumerOptions

        validator.register_consumer(
            RegisterConsumerOptions(
                consumer_id=consumer_id,
                consumer_version=consumer_version,
                schema_id="calculator-api",
                schema_version="1.0.0",
                environment=environment,
                used_endpoints=[
                    EndpointUsage(method="GET", path="/add", used_fields=["result"]),
                    EndpointUsage(method="GET", path="/multiply", used_fields=["result"]),
                    EndpointUsage(method="GET", path="/divide", used_fields=["result"]),
                ],
            )
        )

        consumers = validator.list_consumers("calculator-api", environment)

        assert consumers is not None
        assert isinstance(consumers, list)
        assert any(c["consumer_id"] == consumer_id for c in consumers)


class TestBreakingChangeDetection:

    def test_can_i_deploy_check(
        self, validator, consumer_id, consumer_version, environment
    ):
        from cvt_sdk import RegisterConsumerOptions

        validator.register_consumer(
            RegisterConsumerOptions(
                consumer_id=consumer_id,
                consumer_version=consumer_version,
                schema_id="calculator-api",
                schema_version="1.0.0",
                environment=environment,
                used_endpoints=[
                    EndpointUsage(method="GET", path="/add", used_fields=["result"]),
                    EndpointUsage(method="GET", path="/multiply", used_fields=["result"]),
                    EndpointUsage(method="GET", path="/divide", used_fields=["result"]),
                ],
            )
        )

        result = validator.can_i_deploy("calculator-api", "1.0.0", environment)

        assert result is not None
        assert "safe_to_deploy" in result

"""
HTTP ADAPTER APPROACH
---------------------
Wraps your HTTP client to automatically validate all requests/responses against the schema.

Prerequisites:
- CVT server running
- Producer service running

When to use:
- Integration tests against a real or staging API
- Catch contract violations automatically during existing tests
- Minimal code changes to add validation to existing test suites
"""

import pytest

from cvt_sdk.adapters import ContractValidatingSession


class TestHTTPAdapter:

    @pytest.fixture(scope="class")
    def session(self, validator):
        validation_errors = []

        def on_failure(result, req, resp):
            validation_errors.append({"result": result, "request": req, "response": resp})

        sess = ContractValidatingSession(
            validator,
            auto_validate=True,
            on_validation_failure=on_failure,
        )
        sess._validation_errors = validation_errors
        yield sess

    def test_add_with_automatic_validation(self, session, producer_url):
        response = session.get(f"{producer_url}/add", params={"x": 5, "y": 3})

        assert response.status_code == 200
        assert response.json()["result"] == 8

        interactions = session.get_interactions()
        assert len(interactions) >= 1

        latest = interactions[-1]
        assert latest.validation_result["valid"] is True

    def test_multiply_with_automatic_validation(self, session, producer_url):
        session.clear_interactions()

        response = session.get(f"{producer_url}/multiply", params={"x": 6, "y": 7})

        assert response.status_code == 200
        assert response.json()["result"] == 42

        interactions = session.get_interactions()
        assert len(interactions) == 1
        assert interactions[0].validation_result["valid"] is True

    def test_divide_with_automatic_validation(self, session, producer_url):
        session.clear_interactions()

        response = session.get(f"{producer_url}/divide", params={"x": 20, "y": 4})

        assert response.status_code == 200
        assert response.json()["result"] == 5.0

        interactions = session.get_interactions()
        assert len(interactions) == 1
        assert interactions[0].validation_result["valid"] is True

    def test_divide_by_zero_error_validation(self, session, producer_url):
        session.clear_interactions()

        response = session.get(f"{producer_url}/divide", params={"x": 10, "y": 0})

        assert response.status_code == 400
        assert "error" in response.json()

        interactions = session.get_interactions()
        assert len(interactions) == 1
        assert interactions[0].validation_result["valid"] is True

    def test_captures_request_response_details(self, session, producer_url):
        session.clear_interactions()

        session.get(f"{producer_url}/add", params={"x": 100, "y": 200})

        interactions = session.get_interactions()
        assert len(interactions) == 1

        interaction = interactions[0]
        assert interaction.request["method"] == "GET"
        assert "/add" in interaction.request["path"]
        assert interaction.response["status_code"] == 200
        assert interaction.response["body"]["result"] == 300

    def test_multiple_operations_captured(self, session, producer_url):
        session.clear_interactions()

        session.get(f"{producer_url}/add", params={"x": 1, "y": 2})
        session.get(f"{producer_url}/multiply", params={"x": 3, "y": 4})
        session.get(f"{producer_url}/divide", params={"x": 8, "y": 2})

        interactions = session.get_interactions()
        assert len(interactions) == 3
        assert all(i.validation_result["valid"] for i in interactions)

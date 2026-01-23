"""
MANUAL VALIDATION APPROACH
--------------------------
Explicitly validates request/response pairs by calling validator.validate() directly.

Prerequisites:
- CVT server running
- Producer service running (for tests that make real requests)

When to use:
- Testing specific error scenarios with crafted responses
- Validating responses from non-HTTP sources
- Maximum control over what gets validated and when
"""

import requests


class TestManualValidation:

    def test_add_operation_valid(self, validator, producer_url):
        response = requests.get(f"{producer_url}/add", params={"a": 5, "b": 3})

        request = {
            "method": "GET",
            "path": "/add?a=5&b=3",
            "headers": {},
        }

        validation_response = {
            "status_code": response.status_code,
            "headers": dict(response.headers),
            "body": response.json(),
        }

        result = validator.validate(request, validation_response)

        assert result["valid"] is True
        assert result["errors"] == []
        assert response.json()["result"] == 8

    def test_add_missing_result_field(self, validator):
        request = {
            "method": "GET",
            "path": "/add?a=5&b=3",
            "headers": {},
        }

        invalid_response = {
            "status_code": 200,
            "headers": {"content-type": "application/json"},
            "body": {"total": 8},
        }

        result = validator.validate(request, invalid_response)

        assert result["valid"] is False
        assert len(result["errors"]) > 0

    def test_multiply_operation_valid(self, validator, producer_url):
        response = requests.get(f"{producer_url}/multiply", params={"a": 4, "b": 7})

        request = {
            "method": "GET",
            "path": "/multiply?a=4&b=7",
            "headers": {},
        }

        validation_response = {
            "status_code": response.status_code,
            "headers": dict(response.headers),
            "body": response.json(),
        }

        result = validator.validate(request, validation_response)

        assert result["valid"] is True
        assert result["errors"] == []
        assert response.json()["result"] == 28

    def test_divide_operation_valid(self, validator, producer_url):
        response = requests.get(f"{producer_url}/divide", params={"a": 10, "b": 2})

        request = {
            "method": "GET",
            "path": "/divide?a=10&b=2",
            "headers": {},
        }

        validation_response = {
            "status_code": response.status_code,
            "headers": dict(response.headers),
            "body": response.json(),
        }

        result = validator.validate(request, validation_response)

        assert result["valid"] is True
        assert result["errors"] == []
        assert response.json()["result"] == 5.0

    def test_divide_by_zero_error_response(self, validator, producer_url):
        response = requests.get(f"{producer_url}/divide", params={"a": 10, "b": 0})

        request = {
            "method": "GET",
            "path": "/divide?a=10&b=0",
            "headers": {},
        }

        validation_response = {
            "status_code": response.status_code,
            "headers": dict(response.headers),
            "body": response.json(),
        }

        result = validator.validate(request, validation_response)

        assert result["valid"] is True
        assert response.status_code == 400
        assert "error" in response.json()

    def test_error_response_invalid_structure(self, validator):
        request = {
            "method": "GET",
            "path": "/add?a=invalid&b=3",
            "headers": {},
        }

        invalid_error_response = {
            "status_code": 400,
            "headers": {"content-type": "application/json"},
            "body": {"message": "Invalid input"},
        }

        result = validator.validate(request, invalid_error_response)

        assert result["valid"] is False

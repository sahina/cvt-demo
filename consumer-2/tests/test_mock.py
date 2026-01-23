import pytest

from cvt_sdk.adapters import MockSession


class TestMockClient:

    @pytest.fixture(scope="class")
    def mock_session(self, validator):
        session = MockSession(
            validator,
            cache=True,
            generate_options={"use_examples": True},
        )
        yield session
        session.clear_cache()

    def test_mock_add_response(self, mock_session):
        mock_session.clear_interactions()

        response = mock_session.get("http://calculator-api/add", params={"a": 5, "b": 3})

        assert response.status_code == 200
        data = response.json()
        assert "result" in data
        assert isinstance(data["result"], (int, float))

    def test_mock_multiply_response(self, mock_session):
        mock_session.clear_interactions()

        response = mock_session.get("http://calculator-api/multiply", params={"a": 4, "b": 7})

        assert response.status_code == 200
        data = response.json()
        assert "result" in data
        assert isinstance(data["result"], (int, float))

    def test_mock_divide_response(self, mock_session):
        mock_session.clear_interactions()

        response = mock_session.get("http://calculator-api/divide", params={"a": 10, "b": 2})

        assert response.status_code == 200
        data = response.json()
        assert "result" in data
        assert isinstance(data["result"], (int, float))

    def test_captures_mock_interactions(self, mock_session):
        mock_session.clear_interactions()

        mock_session.get("http://calculator-api/add", params={"a": 10, "b": 20})

        interactions = mock_session.get_interactions()
        assert len(interactions) == 1
        assert interactions[0].request["method"] == "GET"
        assert "/add" in interactions[0].request["path"]

    def test_mock_response_validates_against_schema(self, mock_session, validator):
        mock_session.clear_interactions()

        response = mock_session.get("http://calculator-api/add", params={"a": 1, "b": 1})
        data = response.json()

        validation_result = validator.validate(
            {"method": "GET", "path": "/add?a=1&b=1", "headers": {}},
            {"status_code": 200, "headers": {}, "body": data},
        )

        assert validation_result["valid"] is True

    def test_captures_all_consumer2_endpoints(self, mock_session):
        mock_session.clear_interactions()

        mock_session.get("http://calculator-api/add", params={"a": 5, "b": 3})
        mock_session.get("http://calculator-api/multiply", params={"a": 4, "b": 7})
        mock_session.get("http://calculator-api/divide", params={"a": 10, "b": 2})

        interactions = mock_session.get_interactions()
        assert len(interactions) == 3

        paths = [i.request["path"] for i in interactions]
        assert any("/add" in p for p in paths)
        assert any("/multiply" in p for p in paths)
        assert any("/divide" in p for p in paths)

import os
from pathlib import Path

import pytest

CVT_SERVER_ADDR = os.getenv("CVT_SERVER_ADDR", "localhost:9550")
PRODUCER_URL = os.getenv("PRODUCER_URL", "http://localhost:10001")
SCHEMA_PATH = os.getenv(
    "SCHEMA_PATH", str(Path(__file__).parent.parent.parent / "producer" / "calculator-api.yaml")
)
CONSUMER_ID = "consumer-2"
CONSUMER_VERSION = "1.0.0"
ENVIRONMENT = os.getenv("CVT_ENVIRONMENT", "demo")


@pytest.fixture(scope="module")
def cvt_server_addr():
    return CVT_SERVER_ADDR


@pytest.fixture(scope="module")
def producer_url():
    return PRODUCER_URL


@pytest.fixture(scope="module")
def schema_path():
    return SCHEMA_PATH


@pytest.fixture(scope="module")
def consumer_id():
    return CONSUMER_ID


@pytest.fixture(scope="module")
def consumer_version():
    return CONSUMER_VERSION


@pytest.fixture(scope="module")
def environment():
    return ENVIRONMENT


@pytest.fixture(scope="module")
def validator(cvt_server_addr, schema_path):
    from cvt_sdk import ContractValidator

    v = ContractValidator(cvt_server_addr)
    v.register_schema("calculator-api", schema_path)
    yield v
    v.close()

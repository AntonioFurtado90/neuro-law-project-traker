import json
import urllib.error
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from neurolaw.clients.base_client import HTTPError
from neurolaw.clients.senado_client import SenadoClient

FIXTURES_DIR = Path(__file__).resolve().parents[1] / "fixtures"


def _load_fixture(name: str) -> list:
    return json.loads((FIXTURES_DIR / name).read_text())


def _mock_response(payload) -> MagicMock:
    response = MagicMock()
    response.read.return_value = json.dumps(payload).encode()
    response.__enter__.return_value = response
    return response


def test_fetch_bills_for_year_returns_all_items():
    items = _load_fixture("senado_ano_2026.json")

    with patch("urllib.request.urlopen", side_effect=[_mock_response(items)]) as mock_urlopen:
        client = SenadoClient("https://legis.senado.leg.br/dadosabertos")
        result = client.fetch_bills_for_year(2026)

    assert [item["id"] for item in result] == [item["id"] for item in items]
    assert mock_urlopen.call_count == 1


def test_fetch_bills_for_year_does_not_retry_client_errors():
    http_error = urllib.error.HTTPError(url="x", code=404, msg="not found", hdrs=None, fp=None)

    with patch("urllib.request.urlopen", side_effect=http_error) as mock_urlopen:
        client = SenadoClient("https://legis.senado.leg.br/dadosabertos", max_retries=3)
        with pytest.raises(HTTPError):
            client.fetch_bills_for_year(2026)

    assert mock_urlopen.call_count == 1


def test_fetch_bills_for_year_retries_server_errors_then_succeeds():
    items = _load_fixture("senado_ano_2026.json")
    server_error = urllib.error.HTTPError(url="x", code=503, msg="unavailable", hdrs=None, fp=None)

    with patch(
        "urllib.request.urlopen",
        side_effect=[server_error, _mock_response(items)],
    ) as mock_urlopen, patch("time.sleep") as mock_sleep:
        client = SenadoClient("https://legis.senado.leg.br/dadosabertos", max_retries=3, backoff_seconds=0.01)
        result = client.fetch_bills_for_year(2026)

    assert len(result) == len(items)
    assert mock_urlopen.call_count == 2
    mock_sleep.assert_called_once()

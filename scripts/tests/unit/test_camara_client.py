import json
import urllib.error
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from neurolaw.clients.base_client import HTTPError
from neurolaw.clients.camara_client import CamaraClient

FIXTURES_DIR = Path(__file__).resolve().parents[1] / "fixtures"


def _load_fixture(name: str) -> dict:
    return json.loads((FIXTURES_DIR / name).read_text())


def _mock_response(payload: dict) -> MagicMock:
    response = MagicMock()
    response.read.return_value = json.dumps(payload).encode()
    response.__enter__.return_value = response
    return response


def test_fetch_bills_follows_pagination_across_pages():
    page1 = _load_fixture("camara_page1.json")
    page2 = _load_fixture("camara_page2.json")

    with patch("urllib.request.urlopen", side_effect=[_mock_response(page1), _mock_response(page2)]) as mock_urlopen:
        client = CamaraClient("https://dadosabertos.camara.leg.br/api/v2")
        items = list(client.fetch_bills("2026-08-01", "2026-08-05"))

    assert [item["id"] for item in items] == [2641785, 2641787, 2642095]
    assert mock_urlopen.call_count == 2


def test_fetch_bills_does_not_retry_client_errors():
    http_error = urllib.error.HTTPError(url="x", code=404, msg="not found", hdrs=None, fp=None)

    with patch("urllib.request.urlopen", side_effect=http_error) as mock_urlopen:
        client = CamaraClient("https://dadosabertos.camara.leg.br/api/v2", max_retries=3)
        with pytest.raises(HTTPError):
            list(client.fetch_bills("2026-08-01", "2026-08-05"))

    assert mock_urlopen.call_count == 1


def test_fetch_bills_retries_server_errors_then_succeeds():
    # camara_page2.json has no "next" link, so retry behavior on a single
    # page can be tested without also exercising pagination.
    page2 = _load_fixture("camara_page2.json")
    server_error = urllib.error.HTTPError(url="x", code=503, msg="unavailable", hdrs=None, fp=None)

    with patch(
        "urllib.request.urlopen",
        side_effect=[server_error, _mock_response(page2)],
    ) as mock_urlopen, patch("time.sleep") as mock_sleep:
        client = CamaraClient("https://dadosabertos.camara.leg.br/api/v2", max_retries=3, backoff_seconds=0.01)
        items = list(client.fetch_bills("2026-08-01", "2026-08-05"))

    assert [item["id"] for item in items] == [2642095]
    assert mock_urlopen.call_count == 2
    mock_sleep.assert_called_once()


def test_fetch_bills_raises_after_exhausting_retries():
    server_error = urllib.error.HTTPError(url="x", code=503, msg="unavailable", hdrs=None, fp=None)

    with patch("urllib.request.urlopen", side_effect=server_error) as mock_urlopen, patch("time.sleep"):
        client = CamaraClient("https://dadosabertos.camara.leg.br/api/v2", max_retries=2, backoff_seconds=0.01)
        with pytest.raises(HTTPError):
            list(client.fetch_bills("2026-08-01", "2026-08-05"))

    assert mock_urlopen.call_count == 3

import json
from unittest.mock import patch

import jsonschema

from neurolaw.clients.base_client import HTTPError
from neurolaw.ingestion import ingest_camara

REQUIRED_ENV = {
    "CAMARA_API_BASE_URL": "https://dadosabertos.camara.leg.br/api/v2",
    "RUN_WINDOW_START": "2026-08-01",
    "RUN_WINDOW_END": "2026-08-05",
}

INGESTION_RESULT_SCHEMA = {
    "type": "object",
    "required": ["status", "items", "errors"],
    "properties": {
        "status": {"type": "string", "enum": ["ok", "partial", "failed"]},
        "items": {"type": "array"},
        "errors": {"type": "array"},
    },
}


def _fake_bill(bill_id: str) -> dict:
    return {
        "id": int(bill_id),
        "siglaTipo": "PL",
        "numero": 1,
        "ano": 2026,
        "ementa": "Ementa de teste.",
        "dataApresentacao": "2026-08-02T10:00",
    }


def test_run_writes_ok_envelope_when_fetch_succeeds(tmp_path):
    output_path = tmp_path / "result.json"
    fake_items = [_fake_bill("1"), _fake_bill("2")]

    with patch.dict("os.environ", REQUIRED_ENV, clear=True), patch(
        "neurolaw.ingestion.ingest_camara.CamaraClient"
    ) as mock_client_cls:
        mock_client_cls.return_value.fetch_bills.return_value = iter(fake_items)
        exit_code = ingest_camara.run(str(output_path))

    result = json.loads(output_path.read_text())
    jsonschema.validate(instance=result, schema=INGESTION_RESULT_SCHEMA)
    assert exit_code == 0
    assert result["status"] == "ok"
    assert len(result["items"]) == 2
    assert result["errors"] == []


def test_run_writes_partial_envelope_when_fetch_fails_midway(tmp_path):
    output_path = tmp_path / "result.json"

    def fetch_then_fail():
        yield _fake_bill("1")
        raise HTTPError("page 2 unavailable")

    with patch.dict("os.environ", REQUIRED_ENV, clear=True), patch(
        "neurolaw.ingestion.ingest_camara.CamaraClient"
    ) as mock_client_cls:
        mock_client_cls.return_value.fetch_bills.return_value = fetch_then_fail()
        exit_code = ingest_camara.run(str(output_path))

    result = json.loads(output_path.read_text())
    jsonschema.validate(instance=result, schema=INGESTION_RESULT_SCHEMA)
    assert exit_code == 1
    assert result["status"] == "partial"
    assert len(result["items"]) == 1
    assert len(result["errors"]) == 1


def test_run_writes_failed_envelope_when_first_fetch_fails(tmp_path):
    output_path = tmp_path / "result.json"

    def fail_immediately():
        raise HTTPError("api unavailable")
        yield  # pragma: no cover - makes this a generator

    with patch.dict("os.environ", REQUIRED_ENV, clear=True), patch(
        "neurolaw.ingestion.ingest_camara.CamaraClient"
    ) as mock_client_cls:
        mock_client_cls.return_value.fetch_bills.return_value = fail_immediately()
        exit_code = ingest_camara.run(str(output_path))

    result = json.loads(output_path.read_text())
    jsonschema.validate(instance=result, schema=INGESTION_RESULT_SCHEMA)
    assert exit_code == 1
    assert result["status"] == "failed"
    assert result["items"] == []
    assert len(result["errors"]) == 1

import json
from unittest.mock import patch

import jsonschema

from neurolaw.clients.base_client import HTTPError
from neurolaw.ingestion import ingest_senado

REQUIRED_ENV = {
    "SENADO_API_BASE_URL": "https://legis.senado.leg.br/dadosabertos",
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


def _fake_item(item_id: int, presented_date: str) -> dict:
    return {
        "id": item_id,
        "identificacao": f"PL {item_id}/2026",
        "ementa": "Ementa de teste.",
        "dataApresentacao": presented_date,
        "codigoMateria": item_id,
        "autoria": "Senador(a) Exemplo",
    }


def test_run_writes_ok_envelope_and_filters_by_window(tmp_path):
    output_path = tmp_path / "result.json"
    fake_items = [
        _fake_item(1, "2026-08-03"),  # inside window
        _fake_item(2, "2026-01-06"),  # outside window
    ]

    with patch.dict("os.environ", REQUIRED_ENV, clear=True), patch(
        "neurolaw.ingestion.ingest_senado.SenadoClient"
    ) as mock_client_cls:
        mock_client_cls.return_value.fetch_bills_for_year.return_value = fake_items
        exit_code = ingest_senado.run(str(output_path))

    result = json.loads(output_path.read_text())
    jsonschema.validate(instance=result, schema=INGESTION_RESULT_SCHEMA)
    assert exit_code == 0
    assert result["status"] == "ok"
    assert len(result["items"]) == 1
    assert result["items"][0]["external_id"] == "1"
    assert result["errors"] == []


def test_run_writes_partial_envelope_when_one_year_fails(tmp_path):
    output_path = tmp_path / "result.json"
    env = {**REQUIRED_ENV, "RUN_WINDOW_START": "2025-12-30", "RUN_WINDOW_END": "2026-01-02"}

    def fetch_by_year(year, bill_type="PL"):
        if year == 2025:
            return [_fake_item(1, "2025-12-31")]
        raise HTTPError("2026 unavailable")

    with patch.dict("os.environ", env, clear=True), patch(
        "neurolaw.ingestion.ingest_senado.SenadoClient"
    ) as mock_client_cls:
        mock_client_cls.return_value.fetch_bills_for_year.side_effect = fetch_by_year
        exit_code = ingest_senado.run(str(output_path))

    result = json.loads(output_path.read_text())
    jsonschema.validate(instance=result, schema=INGESTION_RESULT_SCHEMA)
    assert exit_code == 1
    assert result["status"] == "partial"
    assert len(result["items"]) == 1
    assert len(result["errors"]) == 1


def test_run_writes_failed_envelope_when_fetch_fails(tmp_path):
    output_path = tmp_path / "result.json"

    with patch.dict("os.environ", REQUIRED_ENV, clear=True), patch(
        "neurolaw.ingestion.ingest_senado.SenadoClient"
    ) as mock_client_cls:
        mock_client_cls.return_value.fetch_bills_for_year.side_effect = HTTPError("api unavailable")
        exit_code = ingest_senado.run(str(output_path))

    result = json.loads(output_path.read_text())
    jsonschema.validate(instance=result, schema=INGESTION_RESULT_SCHEMA)
    assert exit_code == 1
    assert result["status"] == "failed"
    assert result["items"] == []
    assert len(result["errors"]) == 1

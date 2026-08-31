import json
from pathlib import Path
from unittest.mock import patch

import jsonschema

from neurolaw.relevance import detect_relevance

REPO_ROOT = Path(__file__).resolve().parents[3]
RELEVANCE_SCHEMA = json.loads(
    (REPO_ROOT / "contracts" / "schemas" / "relevance_report.schema.json").read_text()
)
KEYWORD_LIST_PATH = str(
    Path(__file__).resolve().parents[2] / "src" / "neurolaw" / "relevance" / "keyword_list.yaml"
)


def _ingestion_result(items: list[dict]) -> dict:
    return {"status": "ok", "items": items, "errors": []}


def test_run_flags_relevant_and_irrelevant_bills(tmp_path):
    input_path = tmp_path / "ingestion_result.json"
    output_path = tmp_path / "relevance_report.json"
    input_path.write_text(
        json.dumps(
            _ingestion_result(
                [
                    {
                        "source": "camara",
                        "external_id": "1",
                        "ementa": "Institui o Dia Nacional do Pix.",
                    },
                    {
                        "source": "senado",
                        "external_id": "2",
                        "ementa": "Altera a Lei do Fundo Constitucional de Financiamento do Nordeste - FNE.",
                    },
                ]
            )
        )
    )

    with patch.dict("os.environ", {"KEYWORD_LIST_PATH": KEYWORD_LIST_PATH}, clear=True):
        exit_code = detect_relevance.run(str(input_path), str(output_path))

    result = json.loads(output_path.read_text())
    jsonschema.validate(instance=result, schema=RELEVANCE_SCHEMA)
    assert exit_code == 0
    assert result["method"] == "keyword"

    by_id = {r["bill_external_id"]: r for r in result["results"]}
    assert by_id["1"]["is_relevant"] is False
    assert by_id["1"]["matched_keywords"] == []
    assert by_id["2"]["is_relevant"] is True
    assert "FNE" in by_id["2"]["matched_keywords"]

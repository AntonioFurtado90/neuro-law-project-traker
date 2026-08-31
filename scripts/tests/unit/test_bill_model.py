import json
from pathlib import Path

import jsonschema

from neurolaw.models.bill import Bill

REPO_ROOT = Path(__file__).resolve().parents[3]
BILL_SCHEMA = json.loads((REPO_ROOT / "contracts" / "schemas" / "bill.schema.json").read_text())
FIXTURES_DIR = Path(__file__).resolve().parents[1] / "fixtures"


def _load_camara_item(index: int = 0) -> dict:
    page = json.loads((FIXTURES_DIR / "camara_page1.json").read_text())
    return page["dados"][index]


def test_from_camara_item_maps_fields():
    item = _load_camara_item()

    bill = Bill.from_camara_item(item)

    assert bill.source == "camara"
    assert bill.external_id == "2641785"
    assert bill.type == "PL"
    assert bill.number == 4857
    assert bill.year == 2026
    assert bill.ementa == item["ementa"]
    assert bill.presented_date == "2026-08-03"
    assert bill.url == "https://www.camara.leg.br/proposicoesWeb/fichadetramitacao?idProposicao=2641785"
    assert bill.author is None
    assert bill.raw_payload == item


def test_to_dict_conforms_to_bill_schema():
    bill = Bill.from_camara_item(_load_camara_item())

    jsonschema.validate(instance=bill.to_dict(), schema=BILL_SCHEMA)

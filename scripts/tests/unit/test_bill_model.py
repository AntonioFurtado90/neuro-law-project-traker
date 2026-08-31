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


def _load_senado_item(index: int = 0) -> dict:
    items = json.loads((FIXTURES_DIR / "senado_ano_2026.json").read_text())
    return items[index]


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


def test_from_senado_item_maps_fields():
    item = _load_senado_item(0)  # "PL 4874/2026"

    bill = Bill.from_senado_item(item)

    assert bill.source == "senado"
    assert bill.external_id == "9090304"
    assert bill.type == "PL"
    assert bill.number == 4874
    assert bill.year == 2026
    assert bill.ementa == item["ementa"]
    assert bill.presented_date == "2026-08-04"
    assert bill.url == "https://www25.senado.leg.br/web/atividade/materias/-/materia/175341"
    assert bill.author == "Senador Davi Alcolumbre (UNIAO/AP)"
    assert bill.raw_payload == item


def test_from_senado_item_parses_identificacao_with_trailing_text():
    item = _load_senado_item(3)  # "PL 159/2026 (Substitutivo-CD)"

    bill = Bill.from_senado_item(item)

    assert bill.type == "PL"
    assert bill.number == 159
    assert bill.year == 2026


def test_senado_bill_to_dict_conforms_to_bill_schema():
    bill = Bill.from_senado_item(_load_senado_item(0))

    jsonschema.validate(instance=bill.to_dict(), schema=BILL_SCHEMA)

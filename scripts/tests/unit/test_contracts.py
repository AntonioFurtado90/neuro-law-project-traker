import json
from pathlib import Path

import jsonschema
import pytest

REPO_ROOT = Path(__file__).resolve().parents[3]
SCHEMAS_DIR = REPO_ROOT / "contracts" / "schemas"
FIXTURES_DIR = REPO_ROOT / "contracts" / "fixtures"


def _load_schema(name: str) -> dict:
    return json.loads((SCHEMAS_DIR / name).read_text())


@pytest.mark.parametrize(
    "schema_name",
    ["bill.schema.json", "ingestion_result.schema.json", "relevance_report.schema.json"],
)
def test_schema_is_valid_draft7(schema_name):
    schema = _load_schema(schema_name)
    jsonschema.Draft7Validator.check_schema(schema)


def test_sample_bill_fixture_conforms_to_bill_schema():
    schema = _load_schema("bill.schema.json")
    sample = json.loads((FIXTURES_DIR / "sample_bill.json").read_text())

    jsonschema.validate(instance=sample, schema=schema)


def test_ingestion_result_referencing_bill_schema_resolves():
    ingestion_schema = _load_schema("ingestion_result.schema.json")
    bill_schema = _load_schema("bill.schema.json")
    bill_sample = json.loads((FIXTURES_DIR / "sample_bill.json").read_text())
    envelope = {"status": "ok", "items": [bill_sample], "errors": []}

    # Register bill_schema in the resolver's store so the relative $ref
    # resolves locally instead of attempting a network fetch of its $id.
    resolver = jsonschema.RefResolver(
        base_uri=SCHEMAS_DIR.as_uri() + "/",
        referrer=ingestion_schema,
        store={bill_schema["$id"]: bill_schema},
    )
    jsonschema.validate(instance=envelope, schema=ingestion_schema, resolver=resolver)

"""ingest-camara subcommand: fetches bills from the Camara dos Deputados
API for the configured run window and writes an ingestion_result.json
envelope (contracts/schemas/ingestion_result.schema.json)."""

import json
import os

from neurolaw.clients.base_client import HTTPError
from neurolaw.clients.camara_client import CamaraClient
from neurolaw.config import require_env
from neurolaw.models.bill import Bill


def run(output_path: str) -> int:
    base_url = require_env("CAMARA_API_BASE_URL")
    start_date = require_env("RUN_WINDOW_START")
    end_date = require_env("RUN_WINDOW_END")
    bill_type = os.environ.get("CAMARA_BILL_TYPE", "PL")
    max_retries = int(os.environ.get("HTTP_MAX_RETRIES", "3"))
    backoff_seconds = float(os.environ.get("HTTP_BACKOFF_SECONDS", "2"))

    client = CamaraClient(base_url, max_retries=max_retries, backoff_seconds=backoff_seconds)

    items: list[dict] = []
    errors: list[str] = []
    try:
        for raw_item in client.fetch_bills(start_date, end_date, bill_type):
            items.append(Bill.from_camara_item(raw_item).to_dict())
    except HTTPError as exc:
        errors.append(str(exc))

    if errors and not items:
        status = "failed"
    elif errors:
        status = "partial"
    else:
        status = "ok"

    with open(output_path, "w") as f:
        json.dump({"status": status, "items": items, "errors": errors}, f)

    return 0 if status == "ok" else 1

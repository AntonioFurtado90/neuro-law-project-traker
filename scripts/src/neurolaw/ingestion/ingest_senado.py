"""ingest-senado subcommand: fetches bills from the Senado Federal API for
the configured run window and writes an ingestion_result.json envelope
(contracts/schemas/ingestion_result.schema.json).

The Senado "/processo" endpoint only reliably filters by year and bill
type (see clients/senado_client.py), so this fetches every year spanned
by the run window and filters by presentation date client-side.
"""

import json
import os

from neurolaw.clients.base_client import HTTPError
from neurolaw.clients.senado_client import SenadoClient
from neurolaw.config import require_env
from neurolaw.models.bill import Bill


def _years_in_window(start_date: str, end_date: str) -> list[int]:
    start_year = int(start_date[:4])
    end_year = int(end_date[:4])
    return list(range(start_year, end_year + 1))


def run(output_path: str) -> int:
    base_url = require_env("SENADO_API_BASE_URL")
    start_date = require_env("RUN_WINDOW_START")
    end_date = require_env("RUN_WINDOW_END")
    bill_type = os.environ.get("SENADO_BILL_TYPE", "PL")
    max_retries = int(os.environ.get("HTTP_MAX_RETRIES", "3"))
    backoff_seconds = float(os.environ.get("HTTP_BACKOFF_SECONDS", "2"))

    client = SenadoClient(base_url, max_retries=max_retries, backoff_seconds=backoff_seconds)

    items: list[dict] = []
    errors: list[str] = []
    for year in _years_in_window(start_date, end_date):
        try:
            raw_items = client.fetch_bills_for_year(year, bill_type)
        except HTTPError as exc:
            errors.append(str(exc))
            continue

        for raw_item in raw_items:
            presented_date = raw_item.get("dataApresentacao", "")
            if start_date <= presented_date <= end_date:
                items.append(Bill.from_senado_item(raw_item).to_dict())

    if errors and not items:
        status = "failed"
    elif errors:
        status = "partial"
    else:
        status = "ok"

    with open(output_path, "w") as f:
        json.dump({"status": status, "items": items, "errors": errors}, f)

    return 0 if status == "ok" else 1

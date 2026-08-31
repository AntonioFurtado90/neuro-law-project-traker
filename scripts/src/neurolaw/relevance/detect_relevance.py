"""detect-relevance subcommand: reads an ingestion_result.json (as produced
by ingest-camara/ingest-senado) and writes a relevance_report.json
(contracts/schemas/relevance_report.schema.json) flagging bills whose
ementa matches the curated keyword list.
"""

import json

from neurolaw.config import require_env
from neurolaw.relevance.keywords import load_keywords, match


def run(input_path: str, output_path: str) -> int:
    keyword_list_path = require_env("KEYWORD_LIST_PATH")
    keywords = load_keywords(keyword_list_path)

    with open(input_path) as f:
        ingestion_result = json.load(f)

    results = []
    for bill in ingestion_result["items"]:
        matched_keywords = match(bill["ementa"], keywords)
        results.append(
            {
                "bill_external_id": bill["external_id"],
                "source": bill["source"],
                "is_relevant": len(matched_keywords) > 0,
                "matched_keywords": matched_keywords,
                "confidence": None,
            }
        )

    with open(output_path, "w") as f:
        json.dump({"method": "keyword", "model_version": None, "results": results}, f)

    return 0

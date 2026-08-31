"""Client for the Senado Federal "Dados Abertos" API
(https://legis.senado.leg.br/dadosabertos), using the "/processo" endpoint.

The older "/materia/pesquisa/lista" endpoint is officially deprecated
(its own response payload names "/processo" as the replacement) and its
own stated full-deactivation date has already passed, so it is not used
here. "/processo" only reliably filters by "sigla" and "ano" — a
date-range filter does not actually narrow the results — so callers fetch
a whole year and filter by date client-side (see ingestion/ingest_senado.py).
"""

import urllib.parse
from typing import Any

from neurolaw.clients.base_client import fetch_json


class SenadoClient:
    def __init__(self, base_url: str, *, max_retries: int = 3, backoff_seconds: float = 2.0):
        self._base_url = base_url.rstrip("/")
        self._max_retries = max_retries
        self._backoff_seconds = backoff_seconds

    def fetch_bills_for_year(self, year: int, bill_type: str = "PL") -> list[dict[str, Any]]:
        """Returns every raw "processo" item of the given type presented
        in the given year (no further filtering)."""
        query = urllib.parse.urlencode({"sigla": bill_type, "ano": year})
        url = f"{self._base_url}/processo?{query}"
        return fetch_json(url, max_retries=self._max_retries, backoff_seconds=self._backoff_seconds)

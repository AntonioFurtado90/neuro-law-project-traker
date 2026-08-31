"""Client for the Camara dos Deputados "Dados Abertos" API
(https://dadosabertos.camara.leg.br/api/v2)."""

import urllib.parse
from collections.abc import Iterator
from typing import Any

from neurolaw.clients.base_client import fetch_json_pages


class CamaraClient:
    def __init__(self, base_url: str, *, max_retries: int = 3, backoff_seconds: float = 2.0):
        self._base_url = base_url.rstrip("/")
        self._max_retries = max_retries
        self._backoff_seconds = backoff_seconds

    def fetch_bills(self, start_date: str, end_date: str, bill_type: str = "PL") -> Iterator[dict[str, Any]]:
        """Yields raw proposicao items presented between start_date and
        end_date (inclusive, "YYYY-MM-DD"), following pagination.
        """
        query = urllib.parse.urlencode(
            {
                "siglaTipo": bill_type,
                "dataApresentacaoInicio": start_date,
                "dataApresentacaoFim": end_date,
                "itens": 100,
                "ordem": "ASC",
                "ordenarPor": "id",
            }
        )
        start_url = f"{self._base_url}/proposicoes?{query}"

        for page in fetch_json_pages(start_url, max_retries=self._max_retries, backoff_seconds=self._backoff_seconds):
            yield from page.get("dados", [])

"""Shared HTTP helper for the Dados Abertos APIs (Camara/Senado): plain
stdlib urllib is enough for GET + JSON + retry/backoff, so this avoids
adding a runtime dependency (e.g. requests) for something this simple.
"""

import json
import time
import urllib.error
import urllib.request
from collections.abc import Iterator


class HTTPError(RuntimeError):
    pass


def fetch_json_pages(start_url: str, *, max_retries: int, backoff_seconds: float) -> Iterator[dict]:
    """Fetches start_url and follows the response's own "rel": "next" link
    (as returned by the Camara/Senado APIs) until there isn't one, yielding
    each page's decoded JSON body.
    """
    url: str | None = start_url
    while url:
        page = _fetch_json_with_retry(url, max_retries=max_retries, backoff_seconds=backoff_seconds)
        yield page
        url = _next_url(page)


def _fetch_json_with_retry(url: str, *, max_retries: int, backoff_seconds: float) -> dict:
    last_error: Exception | None = None
    for attempt in range(max_retries + 1):
        try:
            with urllib.request.urlopen(url, timeout=30) as response:
                return json.loads(response.read())
        except urllib.error.HTTPError as exc:
            if 400 <= exc.code < 500:
                raise HTTPError(f"client error fetching {url}: HTTP {exc.code}") from exc
            last_error = exc
        except (urllib.error.URLError, TimeoutError) as exc:
            last_error = exc

        if attempt < max_retries:
            time.sleep(backoff_seconds * (attempt + 1))

    raise HTTPError(f"failed to fetch {url} after {max_retries + 1} attempt(s): {last_error}")


def _next_url(page: dict) -> str | None:
    for link in page.get("links", []):
        if link.get("rel") == "next":
            return link.get("href")
    return None

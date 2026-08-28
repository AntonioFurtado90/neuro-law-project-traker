"""All configuration is read from the environment (Twelve-Factor, Factor III).

No defaults are provided for infrastructure values (API URLs, paths) — a
missing required variable is a fatal configuration error, not something to
silently paper over.
"""

import os


class MissingConfigError(RuntimeError):
    pass


def require_env(name: str) -> str:
    value = os.environ.get(name)
    if not value:
        raise MissingConfigError(f"Required environment variable {name!r} is not set.")
    return value

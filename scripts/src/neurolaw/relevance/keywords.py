"""Keyword-based relevance matching (Phase A). See keyword_list.yaml for the
curated term list and docs/roadmap.md for Phase B (ML-based detection)."""

import re

import yaml


def load_keywords(path: str) -> list[str]:
    with open(path) as f:
        return yaml.safe_load(f)


def match(ementa: str, keywords: list[str]) -> list[str]:
    """Returns the subset of keywords that appear in ementa, matched as
    whole words/phrases (word-boundary, case-insensitive) so a short
    acronym like "FNE" doesn't match inside an unrelated word."""
    matched = []
    for keyword in keywords:
        pattern = r"\b" + re.escape(keyword) + r"\b"
        if re.search(pattern, ementa, re.IGNORECASE):
            matched.append(keyword)
    return matched

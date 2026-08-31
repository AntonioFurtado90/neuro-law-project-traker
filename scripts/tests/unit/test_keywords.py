from pathlib import Path

from neurolaw.relevance.keywords import load_keywords, match

KEYWORD_LIST_PATH = (
    Path(__file__).resolve().parents[2] / "src" / "neurolaw" / "relevance" / "keyword_list.yaml"
)


def _real_keywords() -> list[str]:
    return load_keywords(str(KEYWORD_LIST_PATH))


def test_load_keywords_returns_nonempty_list():
    keywords = _real_keywords()

    assert isinstance(keywords, list)
    assert len(keywords) > 0
    assert all(isinstance(k, str) for k in keywords)


def test_match_finds_fundo_constitucional_mention():
    ementa = (
        "Altera a Lei no 7.827, de 1989, que dispoe sobre o Fundo "
        "Constitucional de Financiamento do Norte - FNO."
    )

    matched = match(ementa, _real_keywords())

    assert "Fundo Constitucional de Financiamento" in matched
    assert "FNO" in matched


def test_match_finds_acronym_case_insensitively():
    ementa = "Cria linha de credito no ambito do fco para pequenos produtores."

    matched = match(ementa, _real_keywords())

    assert "FCO" in matched


def test_match_ignores_unrelated_fundo_mentions():
    ementa = "Altera regras do fundo eleitoral e do fundo garantidor de creditos."

    matched = match(ementa, _real_keywords())

    assert matched == []


def test_match_does_not_match_acronym_inside_another_word():
    matched = match("Institui o programa INFONE de capacitacao digital.", ["FNE"])

    assert matched == []

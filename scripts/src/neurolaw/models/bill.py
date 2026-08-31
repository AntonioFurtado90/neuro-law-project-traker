"""Bill mirrors contracts/schemas/bill.schema.json."""

import re
from dataclasses import dataclass, field
from typing import Any

# Matches the leading "<tipo> <numero>/<ano>" of a Senado "identificacao"
# string, e.g. "PL 4874/2026" or "PL 159/2026 (Substitutivo-CD)".
_SENADO_IDENTIFICACAO_RE = re.compile(r"^(?P<type>\S+)\s+(?P<number>\d+)/(?P<year>\d+)")


@dataclass
class Bill:
    source: str
    external_id: str
    type: str
    number: int
    year: int
    ementa: str
    presented_date: str
    url: str
    author: str | None = None
    raw_payload: dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> dict[str, Any]:
        data: dict[str, Any] = {
            "source": self.source,
            "external_id": self.external_id,
            "type": self.type,
            "number": self.number,
            "year": self.year,
            "ementa": self.ementa,
            "presented_date": self.presented_date,
            "url": self.url,
            "raw_payload": self.raw_payload,
        }
        if self.author is not None:
            data["author"] = self.author
        return data

    @classmethod
    def from_camara_item(cls, item: dict[str, Any]) -> "Bill":
        bill_id = item["id"]
        presented_date = item["dataApresentacao"].split("T")[0]
        return cls(
            source="camara",
            external_id=str(bill_id),
            type=item["siglaTipo"],
            number=item["numero"],
            year=item["ano"],
            ementa=item["ementa"],
            presented_date=presented_date,
            url=f"https://www.camara.leg.br/proposicoesWeb/fichadetramitacao?idProposicao={bill_id}",
            raw_payload=item,
        )

    @classmethod
    def from_senado_item(cls, item: dict[str, Any]) -> "Bill":
        identificacao = item["identificacao"]
        match = _SENADO_IDENTIFICACAO_RE.match(identificacao)
        if not match:
            raise ValueError(f"unrecognized Senado identificacao format: {identificacao!r}")

        codigo_materia = item["codigoMateria"]
        return cls(
            source="senado",
            external_id=str(item["id"]),
            type=match.group("type"),
            number=int(match.group("number")),
            year=int(match.group("year")),
            ementa=item["ementa"],
            presented_date=item["dataApresentacao"],
            url=f"https://www25.senado.leg.br/web/atividade/materias/-/materia/{codigo_materia}",
            author=item.get("autoria"),
            raw_payload=item,
        )

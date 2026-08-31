"""Single CLI entrypoint for the neurolaw scripts container.

Subcommands are added incrementally, one per sprint (ingest-camara,
ingest-senado, detect-relevance, ...).
"""

import argparse
import sys

from neurolaw import __version__


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="neurolaw")
    subparsers = parser.add_subparsers(dest="command", required=True)
    subparsers.add_parser("version", help="Print the neurolaw package version.")

    ingest_camara_parser = subparsers.add_parser(
        "ingest-camara", help="Fetch bills from the Camara dos Deputados API."
    )
    ingest_camara_parser.add_argument("--output", required=True, help="Path to write the ingestion_result JSON to.")

    ingest_senado_parser = subparsers.add_parser("ingest-senado", help="Fetch bills from the Senado Federal API.")
    ingest_senado_parser.add_argument("--output", required=True, help="Path to write the ingestion_result JSON to.")

    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)

    if args.command == "version":
        print(__version__)
        return 0

    if args.command == "ingest-camara":
        from neurolaw.ingestion.ingest_camara import run as ingest_camara_run

        return ingest_camara_run(args.output)

    if args.command == "ingest-senado":
        from neurolaw.ingestion.ingest_senado import run as ingest_senado_run

        return ingest_senado_run(args.output)

    parser.error(f"Unknown command: {args.command}")
    return 2


if __name__ == "__main__":
    sys.exit(main())

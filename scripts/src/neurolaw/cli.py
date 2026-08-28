"""Single CLI entrypoint for the neurolaw scripts container.

Subcommands are added incrementally, one per sprint (ingest-camara,
ingest-senado, detect-relevance, ...). Sprint 0 only wires the dispatcher
itself so the container and CI pipeline have something real to build/run.
"""

import argparse
import sys

from neurolaw import __version__


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="neurolaw")
    subparsers = parser.add_subparsers(dest="command", required=True)
    subparsers.add_parser("version", help="Print the neurolaw package version.")
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)

    if args.command == "version":
        print(__version__)
        return 0

    parser.error(f"Unknown command: {args.command}")
    return 2


if __name__ == "__main__":
    sys.exit(main())

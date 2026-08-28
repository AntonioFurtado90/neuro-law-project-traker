from neurolaw import __version__
from neurolaw.cli import main


def test_version_command_prints_package_version(capsys):
    exit_code = main(["version"])

    captured = capsys.readouterr()
    assert exit_code == 0
    assert captured.out.strip() == __version__


def test_missing_command_exits_nonzero():
    try:
        main([])
    except SystemExit as exc:
        assert exc.code != 0
    else:
        raise AssertionError("expected argparse to exit when no command is given")

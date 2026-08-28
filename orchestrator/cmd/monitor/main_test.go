package main

import (
	"bytes"
	"testing"
)

func TestRun_VersionCommand_PrintsVersion(t *testing.T) {
	var out bytes.Buffer

	code := run([]string{"version"}, &out)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := out.String(); got != version+"\n" {
		t.Fatalf("expected %q, got %q", version+"\n", got)
	}
}

func TestRun_UnknownCommand_ReturnsNonZero(t *testing.T) {
	var out bytes.Buffer

	code := run([]string{"bogus"}, &out)

	if code == 0 {
		t.Fatal("expected non-zero exit code for an unknown command")
	}
}

func TestRun_NoArgs_ReturnsNonZero(t *testing.T) {
	var out bytes.Buffer

	code := run([]string{}, &out)

	if code == 0 {
		t.Fatal("expected non-zero exit code when no command is given")
	}
}

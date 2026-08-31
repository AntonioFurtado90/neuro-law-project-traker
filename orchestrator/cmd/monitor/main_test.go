package main

import (
	"bytes"
	"context"
	"testing"
)

func TestRun_VersionCommand_PrintsVersion(t *testing.T) {
	var out bytes.Buffer

	code := run(context.Background(), []string{"version"}, &out)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := out.String(); got != version+"\n" {
		t.Fatalf("expected %q, got %q", version+"\n", got)
	}
}

func TestRun_UnknownCommand_ReturnsNonZero(t *testing.T) {
	var out bytes.Buffer

	code := run(context.Background(), []string{"bogus"}, &out)

	if code == 0 {
		t.Fatal("expected non-zero exit code for an unknown command")
	}
}

func TestRun_NoArgs_ReturnsNonZero(t *testing.T) {
	var out bytes.Buffer

	code := run(context.Background(), []string{}, &out)

	if code == 0 {
		t.Fatal("expected non-zero exit code when no command is given")
	}
}

func TestRun_Migrate_WithoutDatabaseURL_FailsFast(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	var out bytes.Buffer

	code := run(context.Background(), []string{"migrate"}, &out)

	if code == 0 {
		t.Fatal("expected non-zero exit code when DATABASE_URL is unset")
	}
	if out.String() == "" {
		t.Fatal("expected an error message explaining the missing config")
	}
}

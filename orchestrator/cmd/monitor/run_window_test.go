package main

import (
	"bytes"
	"testing"
	"time"
)

func TestComputeRunWindow_MonthRollover(t *testing.T) {
	reference := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)

	start, end := computeRunWindow(reference)

	if start != "2026-08-30" {
		t.Fatalf("expected start 2026-08-30, got %s", start)
	}
	if end != "2026-08-31" {
		t.Fatalf("expected end 2026-08-31, got %s", end)
	}
}

func TestComputeRunWindow_YearRollover(t *testing.T) {
	reference := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	start, end := computeRunWindow(reference)

	if start != "2025-12-30" {
		t.Fatalf("expected start 2025-12-30, got %s", start)
	}
	if end != "2025-12-31" {
		t.Fatalf("expected end 2025-12-31, got %s", end)
	}
}

// computeRunWindow always derives the window from the given reference date,
// with no dependency on any previous run's state — so a "first ever
// execution" needs no special case; this test just makes that explicit.
func TestComputeRunWindow_DoesNotDependOnPriorRuns(t *testing.T) {
	reference := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	firstCallStart, firstCallEnd := computeRunWindow(reference)
	secondCallStart, secondCallEnd := computeRunWindow(reference)

	if firstCallStart != secondCallStart || firstCallEnd != secondCallEnd {
		t.Fatal("expected computeRunWindow to be a pure function of its input")
	}
}

func TestRunRunWindow_WithReferenceDate_PrintsExpectedFormat(t *testing.T) {
	var out bytes.Buffer

	code := runRunWindow([]string{"--reference-date", "2026-09-01"}, &out)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (output: %s)", code, out.String())
	}
	want := "RUN_WINDOW_START=2026-08-30\nRUN_WINDOW_END=2026-08-31\n"
	if out.String() != want {
		t.Fatalf("expected %q, got %q", want, out.String())
	}
}

func TestRunRunWindow_WithInvalidReferenceDate_ReturnsConfigError(t *testing.T) {
	var out bytes.Buffer

	code := runRunWindow([]string{"--reference-date", "not-a-date"}, &out)

	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
}

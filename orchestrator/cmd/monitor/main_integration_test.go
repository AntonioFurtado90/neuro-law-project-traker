//go:build integration

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"neurolaw/orchestrator/internal/db"
)

func repoRootFixture(t *testing.T, relPath string) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine caller for fixture path resolution")
	}
	// this file: orchestrator/cmd/monitor/main_integration_test.go -> repo root is 3 levels up.
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	return filepath.Join(repoRoot, relPath)
}

func TestRun_LoadBills_PersistsBillsFromFixture(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL must be set to run integration tests")
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connecting to test database: %v", err)
	}
	t.Cleanup(pool.Close)
	if _, err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("running migrations: %v", err)
	}

	inputPath := repoRootFixture(t, filepath.Join("contracts", "fixtures", "sample_ingestion_result.json"))
	var out bytes.Buffer

	code := run(ctx, []string{"load-bills", "--input", inputPath}, &out)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (output: %s)", code, out.String())
	}

	var ementa string
	err = pool.QueryRow(ctx,
		`SELECT ementa FROM bills WHERE source = $1 AND external_id = $2`,
		"camara", "2345678",
	).Scan(&ementa)
	if err != nil {
		t.Fatalf("expected the fixture bill to be persisted: %v", err)
	}
	if ementa == "" {
		t.Fatal("expected a non-empty ementa for the persisted bill")
	}
}

func TestRun_GenerateReport_PersistsRelevanceAndWritesReport(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL must be set to run integration tests")
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connecting to test database: %v", err)
	}
	t.Cleanup(pool.Close)
	if _, err := db.Migrate(ctx, pool); err != nil {
		t.Fatalf("running migrations: %v", err)
	}

	billsInputPath := repoRootFixture(t, filepath.Join("contracts", "fixtures", "sample_ingestion_result.json"))
	relevanceInputPath := repoRootFixture(t, filepath.Join("contracts", "fixtures", "sample_relevance_report.json"))
	outputDir := t.TempDir()

	t.Setenv("RUN_WINDOW_END", "2026-01-15")
	t.Setenv("OUTPUT_DIR", outputDir)

	var loadOut bytes.Buffer
	if code := run(ctx, []string{"load-bills", "--input", billsInputPath}, &loadOut); code != 0 {
		t.Fatalf("load-bills failed: exit %d, output: %s", code, loadOut.String())
	}

	var reportOut bytes.Buffer
	code := run(ctx, []string{"generate-report", "--input", relevanceInputPath}, &reportOut)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (output: %s)", code, reportOut.String())
	}

	var isRelevant bool
	err = pool.QueryRow(ctx, `
		SELECT rr.is_relevant FROM relevance_results rr
		JOIN bills b ON b.id = rr.bill_id
		WHERE b.source = $1 AND b.external_id = $2`,
		"camara", "2345678",
	).Scan(&isRelevant)
	if err != nil {
		t.Fatalf("expected a relevance row to be persisted: %v", err)
	}
	if !isRelevant {
		t.Fatal("expected is_relevant to be true")
	}

	reportPath := filepath.Join(outputDir, "report-2026-01-15.md")
	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("expected report file at %s: %v", reportPath, err)
	}
	if !strings.Contains(string(content), "FNO") {
		t.Fatalf("expected report to mention the matched keyword, got: %s", content)
	}

	// reports has no uniqueness constraint (a run can legitimately be
	// reported more than once over time), so this only checks that at
	// least one row was recorded for this run_date, not an exact count —
	// re-running this test against a non-fresh database must not fail.
	var reportCount int
	err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM reports WHERE run_date = $1`, "2026-01-15").Scan(&reportCount)
	if err != nil {
		t.Fatalf("counting reports rows: %v", err)
	}
	if reportCount < 1 {
		t.Fatalf("expected at least 1 reports row for run_date, got %d", reportCount)
	}
}

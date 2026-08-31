//go:build integration

package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
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

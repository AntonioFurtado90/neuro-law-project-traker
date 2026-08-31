//go:build integration

package db

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"neurolaw/orchestrator/internal/models"
)

func setupTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL must be set to run integration tests")
	}

	ctx := context.Background()
	pool, err := Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connecting to test database: %v", err)
	}
	t.Cleanup(pool.Close)

	if _, err := Migrate(ctx, pool); err != nil {
		t.Fatalf("running migrations: %v", err)
	}

	return pool
}

func TestUpsertBill_IsIdempotentOnSourceAndExternalID(t *testing.T) {
	pool := setupTestPool(t)
	repo := NewBillsRepo(pool)
	ctx := context.Background()

	first := models.Bill{
		Source:        "camara",
		ExternalID:    "test-idempotency-1",
		Type:          "PL",
		Number:        1,
		Year:          2026,
		Ementa:        "Ementa original.",
		PresentedDate: "2026-01-01",
		URL:           "https://example.org/1",
	}
	firstID, err := repo.UpsertBill(ctx, first)
	if err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	second := first
	second.Ementa = "Ementa atualizada."
	secondID, err := repo.UpsertBill(ctx, second)
	if err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	if firstID != secondID {
		t.Fatalf("expected same id for repeated (source, external_id), got %d and %d", firstID, secondID)
	}

	var count int
	err = pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM bills WHERE source = $1 AND external_id = $2`,
		first.Source, first.ExternalID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("counting rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 row after two upserts, got %d", count)
	}

	var ementa string
	err = pool.QueryRow(ctx,
		`SELECT ementa FROM bills WHERE id = $1`, firstID,
	).Scan(&ementa)
	if err != nil {
		t.Fatalf("reading back ementa: %v", err)
	}
	if ementa != second.Ementa {
		t.Fatalf("expected ementa to be updated to %q, got %q", second.Ementa, ementa)
	}
}

//go:build integration

package db

import (
	"context"
	"testing"

	"neurolaw/orchestrator/internal/models"
)

func TestRelevanceRepo_RecordRelevance_IsIdempotentPerDay(t *testing.T) {
	pool := setupTestPool(t)
	billsRepo := NewBillsRepo(pool)
	relevanceRepo := NewRelevanceRepo(pool)
	ctx := context.Background()

	bill := models.Bill{
		Source:        "camara",
		ExternalID:    "test-relevance-1",
		Type:          "PL",
		Number:        7,
		Year:          2026,
		Ementa:        "Altera o Fundo Constitucional de Financiamento do Nordeste.",
		PresentedDate: "2026-02-01",
		URL:           "https://example.org/7",
	}
	if _, err := billsRepo.UpsertBill(ctx, bill); err != nil {
		t.Fatalf("upserting bill: %v", err)
	}

	err := relevanceRepo.RecordRelevance(ctx, bill.Source, bill.ExternalID, "keyword", true, []string{"FNE"})
	if err != nil {
		t.Fatalf("first RecordRelevance: %v", err)
	}

	err = relevanceRepo.RecordRelevance(ctx, bill.Source, bill.ExternalID, "keyword", false, []string{})
	if err != nil {
		t.Fatalf("second RecordRelevance: %v", err)
	}

	var count int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM relevance_results rr
		JOIN bills b ON b.id = rr.bill_id
		WHERE b.source = $1 AND b.external_id = $2`,
		bill.Source, bill.ExternalID,
	).Scan(&count)
	if err != nil {
		t.Fatalf("counting rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 relevance row after two calls same day, got %d", count)
	}

	var isRelevant bool
	err = pool.QueryRow(ctx, `
		SELECT rr.is_relevant FROM relevance_results rr
		JOIN bills b ON b.id = rr.bill_id
		WHERE b.source = $1 AND b.external_id = $2`,
		bill.Source, bill.ExternalID,
	).Scan(&isRelevant)
	if err != nil {
		t.Fatalf("reading back is_relevant: %v", err)
	}
	if isRelevant {
		t.Fatal("expected is_relevant to reflect the second (most recent) call")
	}
}

func TestRelevanceRepo_RecordRelevance_ErrorsWhenBillDoesNotExist(t *testing.T) {
	pool := setupTestPool(t)
	relevanceRepo := NewRelevanceRepo(pool)
	ctx := context.Background()

	err := relevanceRepo.RecordRelevance(ctx, "camara", "does-not-exist", "keyword", true, []string{"FNE"})
	if err == nil {
		t.Fatal("expected an error when the referenced bill does not exist")
	}
}

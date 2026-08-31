package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RelevanceRepo persists relevance verdicts, idempotent per (bill, method, day).
type RelevanceRepo struct {
	pool *pgxpool.Pool
}

func NewRelevanceRepo(pool *pgxpool.Pool) *RelevanceRepo {
	return &RelevanceRepo{pool: pool}
}

// RecordRelevance stores a relevance verdict for the bill identified by
// (source, externalID). The bill must already exist (persisted via
// BillsRepo.UpsertBill) — this returns an error otherwise rather than
// silently inserting nothing.
func (r *RelevanceRepo) RecordRelevance(
	ctx context.Context,
	source, externalID, method string,
	isRelevant bool,
	matchedKeywords []string,
) error {
	matchedKeywordsJSON, err := json.Marshal(matchedKeywords)
	if err != nil {
		return fmt.Errorf("marshaling matched_keywords: %w", err)
	}

	tag, err := r.pool.Exec(ctx, `
		INSERT INTO relevance_results (bill_id, method, is_relevant, matched_keywords, evaluated_at, evaluated_date)
		SELECT id, $3, $4, $5, now(), CURRENT_DATE FROM bills WHERE source = $1 AND external_id = $2
		ON CONFLICT (bill_id, method, evaluated_date) DO UPDATE SET
			is_relevant = EXCLUDED.is_relevant,
			matched_keywords = EXCLUDED.matched_keywords,
			evaluated_at = now()`,
		source, externalID, method, isRelevant, matchedKeywordsJSON,
	)
	if err != nil {
		return fmt.Errorf("recording relevance for bill %s/%s: %w", source, externalID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no bill found for %s/%s; run load-bills before generate-report", source, externalID)
	}

	return nil
}

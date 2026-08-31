package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ReportsRepo records the history of what was actually reported.
type ReportsRepo struct {
	pool *pgxpool.Pool
}

func NewReportsRepo(pool *pgxpool.Pool) *ReportsRepo {
	return &ReportsRepo{pool: pool}
}

func (r *ReportsRepo) RecordReport(ctx context.Context, runDate, outputRef string, summary map[string]any) error {
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("marshaling summary: %w", err)
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO reports (run_date, output_ref, summary) VALUES ($1, $2, $3)`,
		runDate, outputRef, summaryJSON,
	)
	if err != nil {
		return fmt.Errorf("recording report for %s: %w", runDate, err)
	}

	return nil
}

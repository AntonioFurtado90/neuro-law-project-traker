package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"neurolaw/orchestrator/internal/models"
)

// BillsRepo persists Bill records, idempotent on (source, external_id).
type BillsRepo struct {
	pool *pgxpool.Pool
}

func NewBillsRepo(pool *pgxpool.Pool) *BillsRepo {
	return &BillsRepo{pool: pool}
}

// UpsertBill inserts a new bill or updates the existing row for the same
// (source, external_id), returning its id either way.
func (r *BillsRepo) UpsertBill(ctx context.Context, bill models.Bill) (int64, error) {
	rawPayload := bill.RawPayload
	if rawPayload == nil {
		rawPayload = map[string]any{}
	}
	rawPayloadJSON, err := json.Marshal(rawPayload)
	if err != nil {
		return 0, fmt.Errorf("marshaling raw_payload: %w", err)
	}

	var id int64
	err = r.pool.QueryRow(ctx, `
		INSERT INTO bills (source, external_id, type, number, year, ementa, author, presented_date, url, raw_payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (source, external_id) DO UPDATE SET
			type = EXCLUDED.type,
			number = EXCLUDED.number,
			year = EXCLUDED.year,
			ementa = EXCLUDED.ementa,
			author = EXCLUDED.author,
			presented_date = EXCLUDED.presented_date,
			url = EXCLUDED.url,
			raw_payload = EXCLUDED.raw_payload,
			last_updated_at = now()
		RETURNING id`,
		bill.Source, bill.ExternalID, bill.Type, bill.Number, bill.Year,
		bill.Ementa, bill.Author, bill.PresentedDate, bill.URL, rawPayloadJSON,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("upserting bill %s/%s: %w", bill.Source, bill.ExternalID, err)
	}

	return id, nil
}

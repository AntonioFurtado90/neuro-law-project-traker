package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

// GetBySourceAndExternalID returns the persisted bill for (source, externalID).
func (r *BillsRepo) GetBySourceAndExternalID(ctx context.Context, source, externalID string) (models.Bill, error) {
	var bill models.Bill
	var presentedDate time.Time
	var rawPayloadJSON []byte

	err := r.pool.QueryRow(ctx, `
		SELECT source, external_id, type, number, year, ementa, author, presented_date, url, raw_payload
		FROM bills WHERE source = $1 AND external_id = $2`,
		source, externalID,
	).Scan(
		&bill.Source, &bill.ExternalID, &bill.Type, &bill.Number, &bill.Year,
		&bill.Ementa, &bill.Author, &presentedDate, &bill.URL, &rawPayloadJSON,
	)
	if err != nil {
		return models.Bill{}, fmt.Errorf("getting bill %s/%s: %w", source, externalID, err)
	}

	bill.PresentedDate = presentedDate.Format("2006-01-02")
	if err := json.Unmarshal(rawPayloadJSON, &bill.RawPayload); err != nil {
		return models.Bill{}, fmt.Errorf("unmarshaling raw_payload for bill %s/%s: %w", source, externalID, err)
	}

	return bill, nil
}

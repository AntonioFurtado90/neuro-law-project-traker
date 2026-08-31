// Package models holds Go types mirroring the shared contracts in
// contracts/schemas/. Field/tag changes here must stay in sync with the
// corresponding schema — see bill_test.go for an automated check.
package models

import "time"

// Bill mirrors contracts/schemas/bill.schema.json.
type Bill struct {
	Source        string         `json:"source"`
	ExternalID    string         `json:"external_id"`
	Type          string         `json:"type"`
	Number        int            `json:"number"`
	Year          int            `json:"year"`
	Ementa        string         `json:"ementa"`
	Author        *string        `json:"author,omitempty"`
	PresentedDate string         `json:"presented_date"`
	URL           string         `json:"url"`
	RawPayload    map[string]any `json:"raw_payload,omitempty"`
}

// PresentedAt parses PresentedDate ("YYYY-MM-DD") into a time.Time.
func (b Bill) PresentedAt() (time.Time, error) {
	return time.Parse("2006-01-02", b.PresentedDate)
}

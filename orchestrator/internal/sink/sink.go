// Package sink turns relevance results into a human-facing report. The
// MVP sink writes a file (see file_sink.go); a future notification channel
// (email/Telegram/Slack, Sprint 6+) implements the same Sink interface
// without requiring any change to the code that calls it.
package sink

import "neurolaw/orchestrator/internal/models"

// RelevantBill pairs a bill with the keywords that flagged it as relevant.
type RelevantBill struct {
	Bill            models.Bill
	MatchedKeywords []string
}

// Report is the input to a Sink: everything needed to produce a
// human-facing summary of one pipeline run.
type Report struct {
	RunDate        string
	TotalEvaluated int
	RelevantBills  []RelevantBill
}

// Sink writes a Report somewhere (a file today, a notification channel
// later) and returns a reference to where it was written (stored in the
// reports table's output_ref column).
type Sink interface {
	Write(report Report) (outputRef string, err error)
}

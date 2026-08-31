package sink

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"neurolaw/orchestrator/internal/models"
)

func TestFileSink_Write_NoRelevantBills(t *testing.T) {
	dir := t.TempDir()
	s := NewFileSink(dir)

	outputRef, err := s.Write(Report{RunDate: "2026-08-05", TotalEvaluated: 3})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	content, err := os.ReadFile(outputRef)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if !strings.Contains(string(content), "Nenhum PL relevante") {
		t.Fatalf("expected a no-relevant-bills message, got: %s", content)
	}
	if filepath.Dir(outputRef) != dir {
		t.Fatalf("expected output in %s, got %s", dir, outputRef)
	}
}

func TestFileSink_Write_WithRelevantBills(t *testing.T) {
	dir := t.TempDir()
	s := NewFileSink(dir)

	report := Report{
		RunDate:        "2026-08-05",
		TotalEvaluated: 2,
		RelevantBills: []RelevantBill{
			{
				Bill: models.Bill{
					Source: "camara", Type: "PL", Number: 1234, Year: 2026,
					Ementa: "Altera o Fundo Constitucional de Financiamento do Norte.",
					URL:    "https://example.org/1234",
				},
				MatchedKeywords: []string{"FNO"},
			},
		},
	}

	outputRef, err := s.Write(report)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	content, err := os.ReadFile(outputRef)
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	got := string(content)
	for _, want := range []string{"PL 1234/2026", "Fundo Constitucional", "FNO", "https://example.org/1234"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected report to contain %q, got: %s", want, got)
		}
	}
}

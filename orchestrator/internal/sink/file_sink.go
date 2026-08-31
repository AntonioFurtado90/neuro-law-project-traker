package sink

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileSink writes the report as a Markdown file under OutputDir.
type FileSink struct {
	OutputDir string
}

func NewFileSink(outputDir string) *FileSink {
	return &FileSink{OutputDir: outputDir}
}

func (s *FileSink) Write(report Report) (string, error) {
	if err := os.MkdirAll(s.OutputDir, 0o755); err != nil {
		return "", fmt.Errorf("creating output dir: %w", err)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# Relatorio de PLs relevantes - %s\n\n", report.RunDate)
	fmt.Fprintf(&b, "Total de PLs avaliados: %d\n", report.TotalEvaluated)
	fmt.Fprintf(&b, "PLs relevantes encontrados: %d\n\n", len(report.RelevantBills))

	if len(report.RelevantBills) == 0 {
		b.WriteString("Nenhum PL relevante encontrado neste periodo.\n")
	} else {
		for _, rb := range report.RelevantBills {
			fmt.Fprintf(&b, "## %s %d/%d (%s)\n\n", rb.Bill.Type, rb.Bill.Number, rb.Bill.Year, rb.Bill.Source)
			fmt.Fprintf(&b, "%s\n\n", rb.Bill.Ementa)
			fmt.Fprintf(&b, "- Termos encontrados: %s\n", strings.Join(rb.MatchedKeywords, ", "))
			fmt.Fprintf(&b, "- Link: %s\n\n", rb.Bill.URL)
		}
	}

	outputPath := filepath.Join(s.OutputDir, fmt.Sprintf("report-%s.md", report.RunDate))
	if err := os.WriteFile(outputPath, []byte(b.String()), 0o644); err != nil {
		return "", fmt.Errorf("writing report file: %w", err)
	}

	return outputPath, nil
}

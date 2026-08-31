// Command monitor is the orchestrator entrypoint. The pipeline (ingest ->
// merge -> detect -> persist -> report) is added incrementally in later
// sprints; today it only wires the command dispatcher, database
// connectivity, migrations, and manual bill persistence.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"neurolaw/orchestrator/internal/config"
	"neurolaw/orchestrator/internal/db"
	"neurolaw/orchestrator/internal/models"
	"neurolaw/orchestrator/internal/sink"
)

const version = "0.1.0"

func run(ctx context.Context, args []string, stdout io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stdout, "usage: monitor <version|migrate|load-bills|generate-report>")
		return 2
	}

	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, version)
		return 0
	case "migrate":
		return runMigrate(ctx, stdout)
	case "load-bills":
		return runLoadBills(ctx, args[1:], stdout)
	case "generate-report":
		return runGenerateReport(ctx, args[1:], stdout)
	default:
		fmt.Fprintf(stdout, "unknown command: %s\n", args[0])
		return 2
	}
}

// connectDB reads DATABASE_URL and opens a pool, writing any error to
// stdout. The returned exit code follows the shared convention: 2 for a
// missing/invalid config, 1 for a connection failure.
func connectDB(ctx context.Context, stdout io.Writer) (*pgxpool.Pool, int) {
	databaseURL, err := config.RequireEnv("DATABASE_URL")
	if err != nil {
		fmt.Fprintln(stdout, err)
		return nil, 2
	}

	pool, err := db.Connect(ctx, databaseURL)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return nil, 1
	}
	return pool, 0
}

func runMigrate(ctx context.Context, stdout io.Writer) int {
	pool, code := connectDB(ctx, stdout)
	if pool == nil {
		return code
	}
	defer pool.Close()

	applied, err := db.Migrate(ctx, pool)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}

	if len(applied) == 0 {
		fmt.Fprintln(stdout, "schema already up to date")
		return 0
	}
	for _, filename := range applied {
		fmt.Fprintf(stdout, "applied %s\n", filename)
	}
	return 0
}

type ingestionResult struct {
	Status string        `json:"status"`
	Items  []models.Bill `json:"items"`
	Errors []string      `json:"errors"`
}

func runLoadBills(ctx context.Context, args []string, stdout io.Writer) int {
	fs := flag.NewFlagSet("load-bills", flag.ContinueOnError)
	fs.SetOutput(stdout)
	inputPath := fs.String("input", "", "path to an ingestion_result.json file")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *inputPath == "" {
		fmt.Fprintln(stdout, "load-bills: --input is required")
		return 2
	}

	raw, err := os.ReadFile(*inputPath)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 2
	}

	var result ingestionResult
	if err := json.Unmarshal(raw, &result); err != nil {
		fmt.Fprintln(stdout, err)
		return 2
	}

	pool, code := connectDB(ctx, stdout)
	if pool == nil {
		return code
	}
	defer pool.Close()

	repo := db.NewBillsRepo(pool)
	var failures int
	for _, bill := range result.Items {
		if _, err := repo.UpsertBill(ctx, bill); err != nil {
			fmt.Fprintln(stdout, err)
			failures++
		}
	}

	fmt.Fprintf(stdout, "persisted %d/%d bill(s)\n", len(result.Items)-failures, len(result.Items))
	if failures > 0 {
		return 1
	}
	return 0
}

// stringSliceFlag implements flag.Value to support a repeatable --input flag.
type stringSliceFlag []string

func (s *stringSliceFlag) String() string { return strings.Join(*s, ",") }
func (s *stringSliceFlag) Set(value string) error {
	*s = append(*s, value)
	return nil
}

type relevanceReport struct {
	Method  string `json:"method"`
	Results []struct {
		BillExternalID  string   `json:"bill_external_id"`
		Source          string   `json:"source"`
		IsRelevant      bool     `json:"is_relevant"`
		MatchedKeywords []string `json:"matched_keywords"`
	} `json:"results"`
}

func runGenerateReport(ctx context.Context, args []string, stdout io.Writer) int {
	fs := flag.NewFlagSet("generate-report", flag.ContinueOnError)
	fs.SetOutput(stdout)
	var inputPaths stringSliceFlag
	fs.Var(&inputPaths, "input", "path to a relevance_report.json file (repeatable)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if len(inputPaths) == 0 {
		fmt.Fprintln(stdout, "generate-report: at least one --input is required")
		return 2
	}

	runDate, err := config.RequireEnv("RUN_WINDOW_END")
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 2
	}
	outputDir, err := config.RequireEnv("OUTPUT_DIR")
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 2
	}

	pool, code := connectDB(ctx, stdout)
	if pool == nil {
		return code
	}
	defer pool.Close()

	relevanceRepo := db.NewRelevanceRepo(pool)
	billsRepo := db.NewBillsRepo(pool)
	reportsRepo := db.NewReportsRepo(pool)

	var totalEvaluated int
	var relevantBills []sink.RelevantBill
	summaryBySource := map[string]int{}

	for _, inputPath := range inputPaths {
		raw, err := os.ReadFile(inputPath)
		if err != nil {
			fmt.Fprintln(stdout, err)
			return 2
		}

		var report relevanceReport
		if err := json.Unmarshal(raw, &report); err != nil {
			fmt.Fprintln(stdout, err)
			return 2
		}

		for _, result := range report.Results {
			totalEvaluated++
			err := relevanceRepo.RecordRelevance(
				ctx, result.Source, result.BillExternalID, report.Method,
				result.IsRelevant, result.MatchedKeywords,
			)
			if err != nil {
				fmt.Fprintln(stdout, err)
				return 1
			}

			if result.IsRelevant {
				bill, err := billsRepo.GetBySourceAndExternalID(ctx, result.Source, result.BillExternalID)
				if err != nil {
					fmt.Fprintln(stdout, err)
					return 1
				}
				relevantBills = append(relevantBills, sink.RelevantBill{Bill: bill, MatchedKeywords: result.MatchedKeywords})
				summaryBySource[result.Source]++
			}
		}
	}

	fileSink := sink.NewFileSink(outputDir)
	outputRef, err := fileSink.Write(sink.Report{
		RunDate:        runDate,
		TotalEvaluated: totalEvaluated,
		RelevantBills:  relevantBills,
	})
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}

	summary := map[string]any{"total_evaluated": totalEvaluated, "relevant_by_source": summaryBySource}
	if err := reportsRepo.RecordReport(ctx, runDate, outputRef, summary); err != nil {
		fmt.Fprintln(stdout, err)
		return 1
	}

	fmt.Fprintf(stdout, "report written to %s (%d relevant of %d evaluated)\n", outputRef, len(relevantBills), totalEvaluated)
	return 0
}

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout))
}

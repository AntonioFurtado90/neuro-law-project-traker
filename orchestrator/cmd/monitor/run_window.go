package main

import (
	"flag"
	"fmt"
	"io"
	"time"
)

// brasiliaLocation is where the run window's default reference date comes
// from: bill presentation dates are meaningful on the Brazilian calendar,
// not UTC. Loading it requires the "time/tzdata" blank import in main.go,
// since the distroless runtime image has no /usr/share/zoneinfo.
const brasiliaLocation = "America/Sao_Paulo"

// computeRunWindow returns a 2-day window ending the day before reference,
// e.g. reference 2026-09-01 -> (2026-08-30, 2026-08-31). The 1-day overlap
// between consecutive daily runs makes the cron resilient to a single
// missed run at no cost: load-bills/RecordRelevance are idempotent, so a
// repeated day is simply a no-op, not a duplicate.
func computeRunWindow(reference time.Time) (startDate, endDate string) {
	end := reference.AddDate(0, 0, -1)
	start := reference.AddDate(0, 0, -2)
	return start.Format("2006-01-02"), end.Format("2006-01-02")
}

func runRunWindow(args []string, stdout io.Writer) int {
	fs := flag.NewFlagSet("run-window", flag.ContinueOnError)
	fs.SetOutput(stdout)
	referenceDateFlag := fs.String(
		"reference-date", "",
		"reference date (YYYY-MM-DD) to compute the window from; defaults to now in "+brasiliaLocation,
	)
	if err := fs.Parse(args); err != nil {
		return 2
	}

	var reference time.Time
	if *referenceDateFlag != "" {
		parsed, err := time.Parse("2006-01-02", *referenceDateFlag)
		if err != nil {
			fmt.Fprintln(stdout, err)
			return 2
		}
		reference = parsed
	} else {
		loc, err := time.LoadLocation(brasiliaLocation)
		if err != nil {
			fmt.Fprintln(stdout, err)
			return 1
		}
		reference = time.Now().In(loc)
	}

	start, end := computeRunWindow(reference)
	fmt.Fprintf(stdout, "RUN_WINDOW_START=%s\n", start)
	fmt.Fprintf(stdout, "RUN_WINDOW_END=%s\n", end)
	return 0
}

// Command monitor is the orchestrator entrypoint. The pipeline (ingest ->
// merge -> detect -> persist -> report) is added incrementally in later
// sprints; today it only wires the command dispatcher, database
// connectivity, and migrations.
package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"neurolaw/orchestrator/internal/config"
	"neurolaw/orchestrator/internal/db"
)

const version = "0.1.0"

func run(ctx context.Context, args []string, stdout io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stdout, "usage: monitor <version|migrate>")
		return 2
	}

	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, version)
		return 0
	case "migrate":
		return runMigrate(ctx, stdout)
	default:
		fmt.Fprintf(stdout, "unknown command: %s\n", args[0])
		return 2
	}
}

func runMigrate(ctx context.Context, stdout io.Writer) int {
	databaseURL, err := config.RequireEnv("DATABASE_URL")
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 2
	}

	pool, err := db.Connect(ctx, databaseURL)
	if err != nil {
		fmt.Fprintln(stdout, err)
		return 1
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

func main() {
	os.Exit(run(context.Background(), os.Args[1:], os.Stdout))
}

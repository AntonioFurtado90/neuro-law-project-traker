// Command monitor is the orchestrator entrypoint. Sprint 0 only wires the
// command dispatcher; the real pipeline (ingest -> merge -> detect -> persist
// -> report) is added incrementally in later sprints.
package main

import (
	"fmt"
	"io"
	"os"
)

const version = "0.1.0"

func run(args []string, stdout io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stdout, "usage: monitor <version>")
		return 2
	}

	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, version)
		return 0
	default:
		fmt.Fprintf(stdout, "unknown command: %s\n", args[0])
		return 2
	}
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout))
}

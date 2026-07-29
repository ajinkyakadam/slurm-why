// Package cli implements datasource.JobDataSource by shelling out to the
// standard Slurm client commands: sacct, squeue, scontrol, and sacctmgr.
//
// These commands are guaranteed to be on $PATH on any login node (a user
// can't submit or manage jobs without them), and every invocation here is
// read-only — no elevated privileges are required.
//
// Output is parsed as parsable2/plain-text rather than --json. Slurm's JSON
// output schema has changed shape across releases (tracking the same
// OpenAPI version skew the REST backend has to negotiate around), whereas
// --parsable2 with an explicit --format field list has been stable for far
// longer and is universally available on the Slurm versions still found on
// login nodes (20.02 through 24.x+). This keeps the CLI backend simple and
// dependency-free at the cost of not using the newer JSON output — a
// reasonable trade for a fallback path that must "just work" everywhere.
package cli

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Runner executes a Slurm client command and returns its stdout. It's an
// injectable seam so tests can substitute fixture output instead of
// invoking real sacct/squeue/scontrol/sacctmgr binaries.
type Runner func(ctx context.Context, name string, args ...string) ([]byte, error)

// execRunner is the default Runner, backed by os/exec.
func execRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return out, nil
}

// Client implements datasource.JobDataSource against the local Slurm CLI.
type Client struct {
	run Runner
}

// New returns a CLI-backed JobDataSource using the real Slurm binaries.
func New() *Client {
	return &Client{run: execRunner}
}

// newWithRunner is used by tests to inject fixture-backed output.
func newWithRunner(run Runner) *Client {
	return &Client{run: run}
}

// Backend identifies this implementation, surfaced via -v/--verbose.
func (c *Client) Backend() string { return "cli" }

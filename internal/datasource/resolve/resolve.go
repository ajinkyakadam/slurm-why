// Package resolve picks a datasource.JobDataSource backend at startup.
//
// It lives in its own package (rather than inside datasource itself)
// because it needs to import both the cli and (later) rest backend
// packages, and those packages import datasource for the shared types —
// putting Resolve in the datasource package itself would create an import
// cycle.
package resolve

import (
	"context"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
	"github.com/ajinkyakadam/slurm-why/internal/datasource/cli"
)

// Resolve returns the JobDataSource this run should use.
//
// For now it always returns the CLI backend — slurmrestd support hasn't
// landed yet. Once it does, this is where the real resolution order goes:
// if a REST endpoint and token are configured, probe it (ping + API version
// negotiation) and use it on success; otherwise fall back to the CLI
// backend silently, logging the choice only under -v/--verbose.
func Resolve(ctx context.Context) datasource.JobDataSource {
	return cli.New()
}

// Package diagnose maps a Slurm job's state onto a plain-English
// Diagnosis (see internal/format), using the read-only JobDataSource
// interface for whatever extra lookups a specific check needs (pending
// reason, partition/QOS ceilings, ...).
package diagnose

import (
	"context"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

// Check is one registered diagnosis.
//
// Applies decides whether Run should even be attempted for a given job.
// Run returns a nil *format.Diagnosis (not an error) when, after fetching
// more data, it determines there's nothing specific to report -- e.g. the
// impossible-request check finding the job's request is in fact within
// every limit it checked. A nil result lets a later, more generic check
// in the registry have a turn instead.
type Check struct {
	ID      string
	Applies func(*datasource.Job) bool
	Run     func(ctx context.Context, ds datasource.JobDataSource, job *datasource.Job) (*format.Diagnosis, error)
}

// Registry is every check slurm-why knows about, in priority order (most
// specific first). Each check registers itself via an init() function in
// its own file, so adding a new check is one new file, not an edit here.
var Registry []Check

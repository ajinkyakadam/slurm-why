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

	// Priority controls run order among checks whose Applies overlap for
	// the same job (lower runs first; default 0). Most checks don't need
	// to set this -- it only matters for a broad catch-all like "pending"
	// that must run after any more specific check has had a chance to
	// defer to it. Diagnose sorts by this explicitly rather than relying
	// on init()/file-registration order, which is easy to get right by
	// accident and silently wrong after a rename or a new file.
	Priority int
}

// Registry is every check slurm-why knows about. Each check registers
// itself via an init() function in its own file, so adding a new check is
// one new file, not an edit here. Run order is decided by Priority (see
// above), not registration order.
var Registry []Check

// pendingRunLast is the Priority a broad PD catch-all check (currently
// just "pending") should use, so any more specific PD check defaults
// (Priority 0) all get a turn first.
const pendingRunLast = 100

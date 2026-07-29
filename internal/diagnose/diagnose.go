package diagnose

import (
	"context"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

// Diagnose runs checks against job in registry order and returns the
// first diagnosis produced. It stops at the first non-nil result so that
// checks with overlapping Applies() conditions don't both fire for the
// same job -- e.g. the impossible-request check takes priority over the
// generic pending-reason explanation, but defers to it (by returning nil)
// when nothing is actually exceeded.
//
// If no check produces a diagnosis, a fallback describing the job's
// current state is returned instead, so a "why" question never comes back
// with silence.
func Diagnose(ctx context.Context, ds datasource.JobDataSource, job *datasource.Job, checks []Check) (format.Diagnosis, error) {
	for _, c := range checks {
		if !c.Applies(job) {
			continue
		}
		d, err := c.Run(ctx, ds, job)
		if err != nil {
			return format.Diagnosis{}, err
		}
		if d != nil {
			return *d, nil
		}
	}
	return fallbackDiagnosis(job), nil
}

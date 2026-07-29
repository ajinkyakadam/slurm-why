package diagnose

import (
	"testing"
	"time"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

func TestFallbackDiagnosis_KnownStates(t *testing.T) {
	states := []string{"COMPLETED", "FAILED", "TIMEOUT", "CANCELLED", "RUNNING", "SOMETHING_UNRECOGNIZED"}
	for _, state := range states {
		job := &datasource.Job{ID: "42", State: state, Elapsed: time.Minute, TimeLimit: time.Hour}
		d := fallbackDiagnosis(job)
		if d.Title == "" || d.Why == "" || d.Fix == "" {
			t.Errorf("fallbackDiagnosis(state=%s) has an empty field: %+v", state, d)
		}
	}
}

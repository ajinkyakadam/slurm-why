package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

func TestRenderTechnicalDetails(t *testing.T) {
	job := &datasource.Job{
		ID:        "42",
		State:     "FAILED",
		User:      "alice",
		Account:   "physics",
		Partition: "debug",
		QOS:       "normal",
		ReqCPUs:   4,
		ReqNodes:  1,
		ReqGPUs:   2,
		Elapsed:   5 * time.Minute,
		TimeLimit: time.Hour,
		ExitCode:  127,
	}
	got := renderTechnicalDetails(job)

	for _, want := range []string{"42", "FAILED", "alice", "physics", "debug", "normal", "cpus=4", "nodes=1", "gpus=2", "code=127"} {
		if !strings.Contains(got, want) {
			t.Errorf("renderTechnicalDetails() = %q, want it to contain %q", got, want)
		}
	}
}

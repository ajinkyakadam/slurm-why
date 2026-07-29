package diagnose

import (
	"context"
	"strings"
	"testing"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

func TestOOMCheck_Applies(t *testing.T) {
	tests := []struct {
		job  datasource.Job
		want bool
	}{
		{datasource.Job{State: "OUT_OF_MEMORY"}, true},
		{datasource.Job{State: "FAILED", ExitSignal: 9}, true},
		{datasource.Job{State: "COMPLETED"}, false},
		{datasource.Job{State: "PENDING"}, false},
	}
	for _, tt := range tests {
		if got := oomApplies(&tt.job); got != tt.want {
			t.Errorf("Applies(%+v) = %v, want %v", tt.job, got, tt.want)
		}
	}
}

func TestRunOOMCheck_UnderSampledUsage(t *testing.T) {
	job := &datasource.Job{
		ID:          "123",
		State:       "OUT_OF_MEMORY",
		ReqMemBytes: 4 * (1 << 30),
		MaxRSSBytes: 4*(1<<30) - (1 << 27), // ~128MiB below the request
	}
	d, err := runOOMCheck(context.Background(), &fakeDataSource{}, job)
	if err != nil {
		t.Fatalf("runOOMCheck: %v", err)
	}
	if !strings.Contains(d.Why, "samples") {
		t.Errorf("Why = %q, want it to explain the sampling-vs-cgroup nuance", d.Why)
	}
	if !strings.Contains(d.Fix, "--mem=") {
		t.Errorf("Fix = %q, want a suggested --mem value", d.Fix)
	}
}

func TestRunOOMCheck_NoUsageRecorded(t *testing.T) {
	job := &datasource.Job{ID: "123", State: "OUT_OF_MEMORY", ReqMemBytes: 4 * (1 << 30)}
	d, err := runOOMCheck(context.Background(), &fakeDataSource{}, job)
	if err != nil {
		t.Fatalf("runOOMCheck: %v", err)
	}
	if strings.Contains(d.Why, "samples") {
		t.Errorf("Why = %q, should not claim under-sampled usage when none was recorded", d.Why)
	}
}

func TestSuggestedMemBytes(t *testing.T) {
	tests := []struct {
		job  datasource.Job
		want int64
	}{
		{datasource.Job{MaxRSSBytes: 1000}, 1250},
		{datasource.Job{ReqMemBytes: 1000}, 1500},
		{datasource.Job{}, 0},
	}
	for _, tt := range tests {
		if got := suggestedMemBytes(&tt.job); got != tt.want {
			t.Errorf("suggestedMemBytes(%+v) = %d, want %d", tt.job, got, tt.want)
		}
	}
}

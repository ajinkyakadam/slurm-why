package diagnose

import (
	"context"
	"testing"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

func TestPendingCheck_Applies(t *testing.T) {
	if !pendingApplies(&datasource.Job{State: "PENDING"}) {
		t.Error("Applies(PENDING) = false, want true")
	}
	if pendingApplies(&datasource.Job{State: "RUNNING"}) {
		t.Error("Applies(RUNNING) = true, want false")
	}
}

func TestRunPendingCheck_KnownReason(t *testing.T) {
	job := &datasource.Job{ID: "5"}
	ds := &fakeDataSource{pendingReason: &datasource.PendingReason{Code: "QOSMaxJobsPerUserLimit"}}

	d, err := runPendingCheck(context.Background(), ds, job)
	if err != nil {
		t.Fatalf("runPendingCheck: %v", err)
	}
	if d.Why != pendingReasons["QOSMaxJobsPerUserLimit"].why {
		t.Errorf("Why = %q, want the QOSMaxJobsPerUserLimit explanation", d.Why)
	}
}

func TestRunPendingCheck_UnknownReason(t *testing.T) {
	job := &datasource.Job{ID: "5"}
	ds := &fakeDataSource{pendingReason: &datasource.PendingReason{Code: "SomeFutureReasonCode"}}

	d, err := runPendingCheck(context.Background(), ds, job)
	if err != nil {
		t.Fatalf("runPendingCheck: %v", err)
	}
	if d.Fix == "" {
		t.Error("expected a generic fallback fix for an unrecognized reason code")
	}
}

func TestRunPendingCheck_PropagatesError(t *testing.T) {
	job := &datasource.Job{ID: "5"}
	ds := &fakeDataSource{pendingErr: errBoom}

	if _, err := runPendingCheck(context.Background(), ds, job); err == nil {
		t.Error("expected an error to propagate from GetPendingReason, got nil")
	}
}

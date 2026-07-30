package diagnose

import (
	"context"
	"strings"
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

func TestRunPendingCheck_AllTop10MiscCodesRecognized(t *testing.T) {
	codes := []string{
		"Reservation", "ReservationInvalid", "ReservationDeleted",
		"PartitionDown", "PartitionInactive", "NodeDown",
		"JobHeldAdmin", "JobHeldUser", "Licenses", "Cleaning",
	}
	job := &datasource.Job{ID: "5"}
	for _, code := range codes {
		ds := &fakeDataSource{pendingReason: &datasource.PendingReason{Code: code}}
		d, err := runPendingCheck(context.Background(), ds, job)
		if err != nil {
			t.Fatalf("runPendingCheck(%s): %v", code, err)
		}
		if d.Why != pendingReasons[code].why {
			t.Errorf("runPendingCheck(%s).Why = %q, want the %s explanation (i.e. code isn't falling through to the generic fallback)", code, d.Why, code)
		}
	}
}

func TestRunPendingCheck_SurfacesDetailSuffix(t *testing.T) {
	job := &datasource.Job{ID: "5"}
	ds := &fakeDataSource{pendingReason: &datasource.PendingReason{
		Code:   "ReqNodeNotAvail",
		Detail: "Reserved for maintenance",
	}}

	d, err := runPendingCheck(context.Background(), ds, job)
	if err != nil {
		t.Fatalf("runPendingCheck: %v", err)
	}
	wantSuffix := ` Slurm adds: "Reserved for maintenance".`
	if !strings.HasSuffix(d.Why, wantSuffix) {
		t.Errorf("Why = %q, want it to end with %q", d.Why, wantSuffix)
	}
}

func TestRunPendingCheck_PropagatesError(t *testing.T) {
	job := &datasource.Job{ID: "5"}
	ds := &fakeDataSource{pendingErr: errBoom}

	if _, err := runPendingCheck(context.Background(), ds, job); err == nil {
		t.Error("expected an error to propagate from GetPendingReason, got nil")
	}
}

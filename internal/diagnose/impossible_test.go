package diagnose

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

func TestImpossibleCheck_Applies(t *testing.T) {
	if !impossibleApplies(&datasource.Job{State: "PENDING"}) {
		t.Error("Applies(PENDING) = false, want true")
	}
	if impossibleApplies(&datasource.Job{State: "COMPLETED"}) {
		t.Error("Applies(COMPLETED) = true, want false")
	}
}

func TestRunImpossibleCheck_DefersOnNonResourcesReason(t *testing.T) {
	job := &datasource.Job{ID: "1", Partition: "compute"}
	ds := &fakeDataSource{pendingReason: &datasource.PendingReason{Code: "Priority"}}

	d, err := runImpossibleCheck(context.Background(), ds, job)
	if err != nil {
		t.Fatalf("runImpossibleCheck: %v", err)
	}
	if d != nil {
		t.Errorf("expected nil diagnosis for a non-Resources reason, got %+v", d)
	}
}

func TestRunImpossibleCheck_DefersWhenWithinLimits(t *testing.T) {
	job := &datasource.Job{ID: "1", Partition: "compute", QOS: "normal", TimeLimit: time.Hour, ReqNodes: 1}
	ds := &fakeDataSource{
		pendingReason: &datasource.PendingReason{Code: "Resources"},
		partitionLimits: map[string]*datasource.PartitionLimits{
			"compute": {MaxTimeSet: true, MaxTimeLimit: 8 * time.Hour, MaxNodes: 10, MaxMemPerNode: datasource.LimitUnlimited},
		},
		qosLimits: map[string]*datasource.QOSLimits{
			"normal": {MaxWallSet: true, MaxWall: 24 * time.Hour},
		},
	}

	d, err := runImpossibleCheck(context.Background(), ds, job)
	if err != nil {
		t.Fatalf("runImpossibleCheck: %v", err)
	}
	if d != nil {
		t.Errorf("expected nil diagnosis when request is within all limits, got %+v", d)
	}
}

func TestRunImpossibleCheck_DetectsTimeLimitViolation(t *testing.T) {
	job := &datasource.Job{ID: "1", Partition: "compute", QOS: "normal", TimeLimit: 12 * time.Hour, ReqNodes: 1}
	ds := &fakeDataSource{
		pendingReason: &datasource.PendingReason{Code: "Resources"},
		partitionLimits: map[string]*datasource.PartitionLimits{
			"compute": {MaxTimeSet: true, MaxTimeLimit: 8 * time.Hour, MaxNodes: 10, MaxMemPerNode: datasource.LimitUnlimited},
		},
		qosLimits: map[string]*datasource.QOSLimits{
			"normal": {},
		},
	}

	d, err := runImpossibleCheck(context.Background(), ds, job)
	if err != nil {
		t.Fatalf("runImpossibleCheck: %v", err)
	}
	if d == nil {
		t.Fatal("expected a diagnosis for a --time request exceeding the partition's MaxTime, got nil")
	}
	if !strings.Contains(d.Why, "partition") {
		t.Errorf("Why = %q, want it to mention the partition time limit violation", d.Why)
	}
}

func TestRunImpossibleCheck_DetectsNodeCountViolation(t *testing.T) {
	job := &datasource.Job{ID: "1", Partition: "compute", QOS: "normal", ReqNodes: 20}
	ds := &fakeDataSource{
		pendingReason: &datasource.PendingReason{Code: "Resources"},
		partitionLimits: map[string]*datasource.PartitionLimits{
			"compute": {MaxNodes: 10, MaxMemPerNode: datasource.LimitUnlimited},
		},
		qosLimits: map[string]*datasource.QOSLimits{"normal": {}},
	}

	d, err := runImpossibleCheck(context.Background(), ds, job)
	if err != nil {
		t.Fatalf("runImpossibleCheck: %v", err)
	}
	if d == nil {
		t.Fatal("expected a diagnosis for a node count exceeding the partition's MaxNodes, got nil")
	}
}

func TestRunImpossibleCheck_PropagatesError(t *testing.T) {
	job := &datasource.Job{ID: "1"}
	ds := &fakeDataSource{pendingErr: errBoom}

	if _, err := runImpossibleCheck(context.Background(), ds, job); err == nil {
		t.Error("expected an error to propagate from GetPendingReason, got nil")
	}
}

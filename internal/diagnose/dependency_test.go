package diagnose

import (
	"context"
	"strings"
	"testing"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

func TestDependencyCheck_Applies(t *testing.T) {
	if !dependencyApplies(&datasource.Job{State: "PENDING"}) {
		t.Error("Applies(PENDING) = false, want true")
	}
	if dependencyApplies(&datasource.Job{State: "COMPLETED"}) {
		t.Error("Applies(COMPLETED) = true, want false")
	}
}

func TestRunDependencyCheck_DefersOnNormalWait(t *testing.T) {
	job := &datasource.Job{ID: "4"}
	ds := &fakeDataSource{
		pendingReason: &datasource.PendingReason{Code: "Dependency"},
		dependency:    &datasource.Dependency{Raw: "afterok:3(unfulfilled)"},
	}

	d, err := runDependencyCheck(context.Background(), ds, job)
	if err != nil {
		t.Fatalf("runDependencyCheck: %v", err)
	}
	if d != nil {
		t.Errorf("expected nil diagnosis for reason=Dependency (normal wait), got %+v", d)
	}
}

func TestRunDependencyCheck_BrokenDependency(t *testing.T) {
	job := &datasource.Job{ID: "2"}
	ds := &fakeDataSource{
		pendingReason: &datasource.PendingReason{Code: "DependencyNeverSatisfied"},
		dependency:    &datasource.Dependency{Raw: "afterok:1(failed)"},
	}

	d, err := runDependencyCheck(context.Background(), ds, job)
	if err != nil {
		t.Fatalf("runDependencyCheck: %v", err)
	}
	if d == nil {
		t.Fatal("expected a diagnosis for reason=DependencyNeverSatisfied, got nil")
	}
	if !strings.Contains(d.Why, "job 1") || !strings.Contains(d.Why, "failed") {
		t.Errorf("Why = %q, want it to name job 1 and its failed outcome", d.Why)
	}
}

func TestRunDependencyCheck_UnparsableSpec(t *testing.T) {
	job := &datasource.Job{ID: "2"}
	ds := &fakeDataSource{
		pendingReason: &datasource.PendingReason{Code: "DependencyNeverSatisfied"},
		dependency:    &datasource.Dependency{Raw: "some-unrecognized-format"},
	}

	d, err := runDependencyCheck(context.Background(), ds, job)
	if err != nil {
		t.Fatalf("runDependencyCheck: %v", err)
	}
	if d == nil {
		t.Fatal("expected a fallback diagnosis even when the spec can't be parsed, got nil")
	}
}

func TestRunDependencyCheck_PropagatesErrors(t *testing.T) {
	job := &datasource.Job{ID: "2"}

	ds := &fakeDataSource{pendingErr: errBoom}
	if _, err := runDependencyCheck(context.Background(), ds, job); err == nil {
		t.Error("expected pendingErr to propagate, got nil")
	}

	ds2 := &fakeDataSource{
		pendingReason: &datasource.PendingReason{Code: "DependencyNeverSatisfied"},
		dependencyErr: errBoom,
	}
	if _, err := runDependencyCheck(context.Background(), ds2, job); err == nil {
		t.Error("expected dependencyErr to propagate, got nil")
	}
}

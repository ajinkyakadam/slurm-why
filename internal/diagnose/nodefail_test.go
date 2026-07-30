package diagnose

import (
	"context"
	"testing"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

func TestNodeFailCheck_Applies(t *testing.T) {
	if !nodeFailApplies(&datasource.Job{State: "NODE_FAIL"}) {
		t.Error("Applies(NODE_FAIL) = false, want true")
	}
	if nodeFailApplies(&datasource.Job{State: "FAILED"}) {
		t.Error("Applies(FAILED) = true, want false")
	}
	if nodeFailApplies(&datasource.Job{State: "PENDING"}) {
		t.Error("Applies(PENDING) = true, want false")
	}
}

func TestRunNodeFailCheck(t *testing.T) {
	job := &datasource.Job{ID: "6", State: "NODE_FAIL"}
	d, err := runNodeFailCheck(context.Background(), &fakeDataSource{}, job)
	if err != nil {
		t.Fatalf("runNodeFailCheck: %v", err)
	}
	if d == nil {
		t.Fatal("expected a diagnosis, got nil")
	}
	if d.Title == "" || d.Why == "" || d.Fix == "" {
		t.Errorf("diagnosis has an empty field: %+v", d)
	}
}

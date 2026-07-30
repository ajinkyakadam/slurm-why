package diagnose

import (
	"context"
	"testing"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

func TestDiagnose_FirstMatchWins(t *testing.T) {
	job := &datasource.Job{ID: "1", State: "PENDING"}
	ds := &fakeDataSource{}

	checks := []Check{
		{ID: "a", Applies: alwaysApplies, Run: func(context.Context, datasource.JobDataSource, *datasource.Job) (*format.Diagnosis, error) {
			return &format.Diagnosis{Title: "first"}, nil
		}},
		{ID: "b", Applies: alwaysApplies, Run: func(context.Context, datasource.JobDataSource, *datasource.Job) (*format.Diagnosis, error) {
			t.Fatal("check b should not run once check a produced a diagnosis")
			return nil, nil
		}},
	}

	got, err := Diagnose(context.Background(), ds, job, checks)
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}
	if got.Title != "first" {
		t.Errorf("Diagnose().Title = %q, want %q", got.Title, "first")
	}
}

func TestDiagnose_RunsInPriorityOrderNotRegistrationOrder(t *testing.T) {
	job := &datasource.Job{ID: "1", State: "PENDING"}
	ds := &fakeDataSource{}

	// Registered out of intended run order (catch-all first, specific
	// check second) -- Priority must be what decides run order, not
	// slice position.
	checks := []Check{
		{ID: "catch-all", Priority: pendingRunLast, Applies: alwaysApplies, Run: func(context.Context, datasource.JobDataSource, *datasource.Job) (*format.Diagnosis, error) {
			return &format.Diagnosis{Title: "generic"}, nil
		}},
		{ID: "specific", Priority: 0, Applies: alwaysApplies, Run: func(context.Context, datasource.JobDataSource, *datasource.Job) (*format.Diagnosis, error) {
			return &format.Diagnosis{Title: "specific"}, nil
		}},
	}

	got, err := Diagnose(context.Background(), ds, job, checks)
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}
	if got.Title != "specific" {
		t.Errorf("Diagnose().Title = %q, want %q (lower Priority should run first regardless of slice order)", got.Title, "specific")
	}
}

func TestDiagnose_DefersWhenCheckReturnsNil(t *testing.T) {
	job := &datasource.Job{ID: "1", State: "PENDING"}
	ds := &fakeDataSource{}

	checks := []Check{
		{ID: "a", Applies: alwaysApplies, Run: func(context.Context, datasource.JobDataSource, *datasource.Job) (*format.Diagnosis, error) {
			return nil, nil // defers to the next check
		}},
		{ID: "b", Applies: alwaysApplies, Run: func(context.Context, datasource.JobDataSource, *datasource.Job) (*format.Diagnosis, error) {
			return &format.Diagnosis{Title: "second"}, nil
		}},
	}

	got, err := Diagnose(context.Background(), ds, job, checks)
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}
	if got.Title != "second" {
		t.Errorf("Diagnose().Title = %q, want %q", got.Title, "second")
	}
}

func TestDiagnose_FallsBackWhenNoCheckApplies(t *testing.T) {
	job := &datasource.Job{ID: "1", State: "COMPLETED", ExitCode: 0}
	ds := &fakeDataSource{}

	got, err := Diagnose(context.Background(), ds, job, nil)
	if err != nil {
		t.Fatalf("Diagnose: %v", err)
	}
	if got.Title != "Job completed successfully" {
		t.Errorf("Diagnose().Title = %q, want the fallback completed message", got.Title)
	}
}

func alwaysApplies(*datasource.Job) bool { return true }

func TestDiagnoseWithCheckID_ReturnsFiredCheckID(t *testing.T) {
	job := &datasource.Job{ID: "1", State: "PENDING"}
	ds := &fakeDataSource{}

	checks := []Check{
		{ID: "a", Applies: alwaysApplies, Run: func(context.Context, datasource.JobDataSource, *datasource.Job) (*format.Diagnosis, error) {
			return nil, nil
		}},
		{ID: "b", Applies: alwaysApplies, Run: func(context.Context, datasource.JobDataSource, *datasource.Job) (*format.Diagnosis, error) {
			return &format.Diagnosis{Title: "second"}, nil
		}},
	}

	_, checkID, err := DiagnoseWithCheckID(context.Background(), ds, job, checks)
	if err != nil {
		t.Fatalf("DiagnoseWithCheckID: %v", err)
	}
	if checkID != "b" {
		t.Errorf("checkID = %q, want %q", checkID, "b")
	}
}

func TestDiagnoseWithCheckID_EmptyWhenFallbackFires(t *testing.T) {
	job := &datasource.Job{ID: "1", State: "COMPLETED"}
	ds := &fakeDataSource{}

	_, checkID, err := DiagnoseWithCheckID(context.Background(), ds, job, nil)
	if err != nil {
		t.Fatalf("DiagnoseWithCheckID: %v", err)
	}
	if checkID != "" {
		t.Errorf("checkID = %q, want empty when no check applied", checkID)
	}
}

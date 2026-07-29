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

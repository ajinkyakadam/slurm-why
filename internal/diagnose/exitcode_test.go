package diagnose

import (
	"context"
	"strings"
	"testing"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

func TestExitCodeCheck_Applies(t *testing.T) {
	tests := []struct {
		job  datasource.Job
		want bool
	}{
		{datasource.Job{State: "FAILED", ExitCode: 127}, true},
		{datasource.Job{State: "FAILED", ExitCode: 137}, false},   // owned by oom
		{datasource.Job{State: "FAILED", ExitSignal: 9}, false},   // owned by oom
		{datasource.Job{State: "FAILED", ExitSignal: 125}, false}, // owned by oom
		{datasource.Job{State: "COMPLETED", ExitCode: 127}, false},
	}
	for _, tt := range tests {
		if got := exitCodeApplies(&tt.job); got != tt.want {
			t.Errorf("Applies(%+v) = %v, want %v", tt.job, got, tt.want)
		}
	}
}

func TestRunExitCodeCheck_CommandNotFound(t *testing.T) {
	job := &datasource.Job{ExitCode: 127}
	d, err := runExitCodeCheck(context.Background(), &fakeDataSource{}, job)
	if err != nil {
		t.Fatalf("runExitCodeCheck: %v", err)
	}
	if d == nil || !strings.Contains(d.Why, "command not found") {
		t.Fatalf("expected a command-not-found diagnosis, got %+v", d)
	}
}

func TestRunExitCodeCheck_NotExecutable(t *testing.T) {
	job := &datasource.Job{ExitCode: 126}
	d, err := runExitCodeCheck(context.Background(), &fakeDataSource{}, job)
	if err != nil {
		t.Fatalf("runExitCodeCheck: %v", err)
	}
	if d == nil || !strings.Contains(d.Why, "126") {
		t.Fatalf("expected a not-executable diagnosis, got %+v", d)
	}
}

func TestRunExitCodeCheck_SegfaultViaExitCode(t *testing.T) {
	job := &datasource.Job{ExitCode: 139}
	d, err := runExitCodeCheck(context.Background(), &fakeDataSource{}, job)
	if err != nil {
		t.Fatalf("runExitCodeCheck: %v", err)
	}
	if d == nil || !strings.Contains(d.Why, "SIGSEGV") {
		t.Fatalf("expected a segfault diagnosis, got %+v", d)
	}
}

func TestRunExitCodeCheck_SegfaultViaExitSignal(t *testing.T) {
	job := &datasource.Job{ExitCode: 0, ExitSignal: 11}
	d, err := runExitCodeCheck(context.Background(), &fakeDataSource{}, job)
	if err != nil {
		t.Fatalf("runExitCodeCheck: %v", err)
	}
	if d == nil {
		t.Fatal("expected a segfault diagnosis for ExitSignal=11, got nil")
	}
	if strings.Contains(d.Title, "exit code 0") {
		t.Errorf("Title = %q, should not misleadingly cite exit code 0 for the signal-based case", d.Title)
	}
}

func TestRunExitCodeCheck_UnrecognizedDefers(t *testing.T) {
	job := &datasource.Job{ExitCode: 1}
	d, err := runExitCodeCheck(context.Background(), &fakeDataSource{}, job)
	if err != nil {
		t.Fatalf("runExitCodeCheck: %v", err)
	}
	if d != nil {
		t.Errorf("expected nil for an unrecognized exit code, got %+v", d)
	}
}

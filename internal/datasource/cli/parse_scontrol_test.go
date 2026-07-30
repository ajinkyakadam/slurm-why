package cli

import (
	"context"
	"testing"
	"time"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

func TestParsePartitionOutput_Unlimited(t *testing.T) {
	out := readFixture(t, "scontrol_partition_compute.txt")
	p, err := parsePartitionOutput(out, "compute")
	if err != nil {
		t.Fatalf("parsePartitionOutput: %v", err)
	}

	if p.Name != "compute" {
		t.Errorf("Name = %q, want compute", p.Name)
	}
	if !p.MaxTimeSet || p.MaxTimeLimit != 8*time.Hour {
		t.Errorf("MaxTimeLimit = %v (set=%v), want 8h (set=true)", p.MaxTimeLimit, p.MaxTimeSet)
	}
	if p.MaxNodes != 10 {
		t.Errorf("MaxNodes = %d, want 10", p.MaxNodes)
	}
	if p.MaxMemPerNode != datasource.LimitUnlimited {
		t.Errorf("MaxMemPerNode = %d, want LimitUnlimited", p.MaxMemPerNode)
	}
}

func TestParsePartitionOutput_NumericMemLimit(t *testing.T) {
	out := readFixture(t, "scontrol_partition_gpu.txt")
	p, err := parsePartitionOutput(out, "gpu")
	if err != nil {
		t.Fatalf("parsePartitionOutput: %v", err)
	}

	wantBytes := int64(131072) * (1 << 20)
	if p.MaxMemPerNode != wantBytes {
		t.Errorf("MaxMemPerNode = %d, want %d", p.MaxMemPerNode, wantBytes)
	}
	if p.MaxNodes != 2 {
		t.Errorf("MaxNodes = %d, want 2", p.MaxNodes)
	}
}

func TestParsePartitionOutput_NotFound(t *testing.T) {
	if _, err := parsePartitionOutput([]byte(""), "ghost"); err == nil {
		t.Fatal("expected an error for a nonexistent partition, got nil")
	}
}

func TestParseDependencyOutput_Failed(t *testing.T) {
	out := readFixture(t, "scontrol_job_dependency_failed.txt")
	d, err := parseDependencyOutput(out, "2")
	if err != nil {
		t.Fatalf("parseDependencyOutput: %v", err)
	}
	// Confirms the parser isn't fooled by the lowercase "--dependency=..."
	// embedded in SubmitLine -- it must pick up the real, capitalized
	// top-level Dependency field.
	if d.Raw != "afterok:1(failed)" {
		t.Errorf("Raw = %q, want %q", d.Raw, "afterok:1(failed)")
	}
}

func TestParseDependencyOutput_Unfulfilled(t *testing.T) {
	out := readFixture(t, "scontrol_job_dependency_unfulfilled.txt")
	d, err := parseDependencyOutput(out, "4")
	if err != nil {
		t.Fatalf("parseDependencyOutput: %v", err)
	}
	if d.Raw != "afterok:3(unfulfilled)" {
		t.Errorf("Raw = %q, want %q", d.Raw, "afterok:3(unfulfilled)")
	}
}

func TestParseDependencyOutput_None(t *testing.T) {
	out := readFixture(t, "scontrol_job_no_dependency.txt")
	d, err := parseDependencyOutput(out, "1")
	if err != nil {
		t.Fatalf("parseDependencyOutput: %v", err)
	}
	if d.Raw != "" {
		t.Errorf("Raw = %q, want empty", d.Raw)
	}
}

func TestParseDependencyOutput_NotFound(t *testing.T) {
	if _, err := parseDependencyOutput([]byte(""), "999"); err == nil {
		t.Fatal("expected an error for a nonexistent job, got nil")
	}
}

func TestParseReservationOutput_Inactive(t *testing.T) {
	out := readFixture(t, "scontrol_reservation_inactive.txt")
	r, err := parseReservationOutput(out, "maint5")
	if err != nil {
		t.Fatalf("parseReservationOutput: %v", err)
	}
	if r.Name != "maint5" {
		t.Errorf("Name = %q, want maint5", r.Name)
	}
	if r.State != "INACTIVE" {
		t.Errorf("State = %q, want INACTIVE", r.State)
	}
	wantStart := time.Date(2026, 7, 30, 8, 0, 0, 0, time.UTC)
	if !r.StartTime.Equal(wantStart) {
		t.Errorf("StartTime = %v, want %v", r.StartTime, wantStart)
	}
	if r.Nodes != "slurmtest" {
		t.Errorf("Nodes = %q, want slurmtest", r.Nodes)
	}
}

func TestParseReservationOutput_Active(t *testing.T) {
	out := readFixture(t, "scontrol_reservation_active.txt")
	r, err := parseReservationOutput(out, "maint4")
	if err != nil {
		t.Fatalf("parseReservationOutput: %v", err)
	}
	if r.State != "ACTIVE" {
		t.Errorf("State = %q, want ACTIVE", r.State)
	}
}

func TestParseReservationOutput_NotFound(t *testing.T) {
	if _, err := parseReservationOutput([]byte(""), "ghost"); err == nil {
		t.Fatal("expected an error for a nonexistent reservation, got nil")
	}
}

func TestClient_GetReservation_JobHasReservation(t *testing.T) {
	c := newWithRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		switch args[1] {
		case "job":
			return readFixture(t, "scontrol_job_reservation_pending.txt"), nil
		case "reservation":
			return readFixture(t, "scontrol_reservation_inactive.txt"), nil
		default:
			t.Fatalf("unexpected scontrol subcommand %q", args[1])
			return nil, nil
		}
	})

	r, err := c.GetReservation(context.Background(), "44")
	if err != nil {
		t.Fatalf("GetReservation: %v", err)
	}
	if r == nil {
		t.Fatal("GetReservation = nil, want a reservation")
	}
	if r.Name != "maint5" {
		t.Errorf("Name = %q, want maint5", r.Name)
	}
}

func TestClient_GetReservation_JobHasNoReservation(t *testing.T) {
	c := newWithRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return readFixture(t, "scontrol_job_no_reservation.txt"), nil
	})

	r, err := c.GetReservation(context.Background(), "36")
	if err != nil {
		t.Fatalf("GetReservation: %v", err)
	}
	if r != nil {
		t.Errorf("GetReservation = %+v, want nil (job doesn't reference a reservation)", r)
	}
}

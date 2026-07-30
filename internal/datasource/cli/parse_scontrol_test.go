package cli

import (
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

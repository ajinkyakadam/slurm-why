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

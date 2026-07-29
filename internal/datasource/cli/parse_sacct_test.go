package cli

import (
	"os"
	"testing"
	"time"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("reading fixture %s: %v", name, err)
	}
	return data
}

func TestParseSacctOutput_OOM(t *testing.T) {
	out := readFixture(t, "sacct_oom.txt")
	job, err := parseSacctOutput(out, "123")
	if err != nil {
		t.Fatalf("parseSacctOutput: %v", err)
	}

	if job.State != "OUT_OF_MEMORY" {
		t.Errorf("State = %q, want OUT_OF_MEMORY", job.State)
	}
	if job.ExitCode != 0 || job.ExitSignal != 125 {
		t.Errorf("ExitCode/Signal = %d/%d, want 0/125", job.ExitCode, job.ExitSignal)
	}
	wantReqMem := int64(4) * 1024 * 1024 * 1024
	if job.ReqMemBytes != wantReqMem {
		t.Errorf("ReqMemBytes = %d, want %d", job.ReqMemBytes, wantReqMem)
	}
	// MaxRSS must be the max across all steps (the .batch step here),
	// not the (blank) parent job line — and it should land *below*
	// ReqMem, demonstrating the OOM-can-fire-under-sampled-usage case.
	wantMaxRSS := int64(4123456) * 1024
	if job.MaxRSSBytes != wantMaxRSS {
		t.Errorf("MaxRSSBytes = %d, want %d", job.MaxRSSBytes, wantMaxRSS)
	}
	if job.MaxRSSBytes >= job.ReqMemBytes {
		t.Errorf("fixture should demonstrate MaxRSS (%d) < ReqMem (%d)", job.MaxRSSBytes, job.ReqMemBytes)
	}
	if job.Elapsed != 12*time.Minute+34*time.Second {
		t.Errorf("Elapsed = %v, want 12m34s", job.Elapsed)
	}
	if job.TimeLimit != time.Hour {
		t.Errorf("TimeLimit = %v, want 1h", job.TimeLimit)
	}
	if job.ReqCPUs != 2 || job.ReqNodes != 1 || job.ReqGPUs != 0 {
		t.Errorf("Req(CPUs,Nodes,GPUs) = (%d,%d,%d), want (2,1,0)", job.ReqCPUs, job.ReqNodes, job.ReqGPUs)
	}
}

func TestParseSacctOutput_Normal(t *testing.T) {
	out := readFixture(t, "sacct_normal.txt")
	job, err := parseSacctOutput(out, "456")
	if err != nil {
		t.Fatalf("parseSacctOutput: %v", err)
	}

	if job.State != "COMPLETED" {
		t.Errorf("State = %q, want COMPLETED", job.State)
	}
	if job.ExitCode != 0 || job.ExitSignal != 0 {
		t.Errorf("ExitCode/Signal = %d/%d, want 0/0", job.ExitCode, job.ExitSignal)
	}
	if job.ReqCPUs != 4 {
		t.Errorf("ReqCPUs = %d, want 4", job.ReqCPUs)
	}
}

func TestParseSacctOutput_NoRecord(t *testing.T) {
	if _, err := parseSacctOutput([]byte(""), "999"); err == nil {
		t.Fatal("expected an error for empty sacct output, got nil")
	}
}

func TestParseSacctExitCode(t *testing.T) {
	tests := []struct {
		in         string
		wantCode   int
		wantSignal int
	}{
		{"0:0", 0, 0},
		{"137:9", 137, 9},
		{"1", 1, 0},
	}
	for _, tt := range tests {
		code, signal := parseSacctExitCode(tt.in)
		if code != tt.wantCode || signal != tt.wantSignal {
			t.Errorf("parseSacctExitCode(%q) = (%d,%d), want (%d,%d)", tt.in, code, signal, tt.wantCode, tt.wantSignal)
		}
	}
}

package cli

import "testing"

func TestParseSqueueOutput(t *testing.T) {
	tests := []struct {
		fixture  string
		jobID    string
		wantCode string
	}{
		{"squeue_pending_priority.txt", "789", "Priority"},
		{"squeue_pending_qosmax.txt", "790", "QOSMaxJobsPerUserLimit"},
	}
	for _, tt := range tests {
		out := readFixture(t, tt.fixture)
		reason, err := parseSqueueOutput(out, tt.jobID)
		if err != nil {
			t.Fatalf("parseSqueueOutput(%s): %v", tt.fixture, err)
		}
		if reason.Code != tt.wantCode {
			t.Errorf("parseSqueueOutput(%s).Code = %q, want %q", tt.fixture, reason.Code, tt.wantCode)
		}
	}
}

func TestParseSqueueOutput_NotFound(t *testing.T) {
	if _, err := parseSqueueOutput([]byte(""), "999"); err == nil {
		t.Fatal("expected an error when squeue has no rows for the job, got nil")
	}
}

func TestParsePendingJobIDs(t *testing.T) {
	out := readFixture(t, "squeue_pending_jobs.txt")
	got := parsePendingJobIDs(out)
	want := []string{"10", "23", "57"}
	if len(got) != len(want) {
		t.Fatalf("parsePendingJobIDs() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("parsePendingJobIDs()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParsePendingJobIDs_Empty(t *testing.T) {
	if got := parsePendingJobIDs([]byte("")); len(got) != 0 {
		t.Errorf("parsePendingJobIDs(empty) = %v, want empty", got)
	}
}

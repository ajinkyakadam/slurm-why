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

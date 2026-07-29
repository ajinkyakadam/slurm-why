package cli

import (
	"context"
	"testing"
)

// TestClient_GetJob_InjectedRunner exercises Client.GetJob end-to-end
// through the Runner seam, proving no real sacct binary is invoked: the
// fake runner asserts on the exact command line and returns fixture bytes.
func TestClient_GetJob_InjectedRunner(t *testing.T) {
	fixture := readFixture(t, "sacct_oom.txt")

	var gotName string
	var gotArgs []string
	fake := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		gotName = name
		gotArgs = args
		return fixture, nil
	}

	c := newWithRunner(fake)
	if c.Backend() != "cli" {
		t.Errorf("Backend() = %q, want cli", c.Backend())
	}

	job, err := c.GetJob(context.Background(), "123")
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if job.State != "OUT_OF_MEMORY" {
		t.Errorf("State = %q, want OUT_OF_MEMORY", job.State)
	}
	if gotName != "sacct" {
		t.Errorf("runner invoked with name %q, want sacct", gotName)
	}
	if len(gotArgs) == 0 || gotArgs[0] != "-j" || gotArgs[1] != "123" {
		t.Errorf("runner args = %v, want to start with [-j 123]", gotArgs)
	}
}

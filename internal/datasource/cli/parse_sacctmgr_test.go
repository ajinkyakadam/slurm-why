package cli

import (
	"testing"
	"time"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

func TestParseQOSOutput_Normal(t *testing.T) {
	out := readFixture(t, "sacctmgr_qos_normal.txt")
	q, err := parseQOSOutput(out, "normal")
	if err != nil {
		t.Fatalf("parseQOSOutput: %v", err)
	}
	if q.MaxJobsPerUser != 4 {
		t.Errorf("MaxJobsPerUser = %d, want 4", q.MaxJobsPerUser)
	}
	if q.MaxSubmitPerUser != 8 {
		t.Errorf("MaxSubmitPerUser = %d, want 8", q.MaxSubmitPerUser)
	}
	if !q.MaxWallSet || q.MaxWall != 24*time.Hour {
		t.Errorf("MaxWall = %v (set=%v), want 24h (set=true)", q.MaxWall, q.MaxWallSet)
	}
}

func TestParseQOSOutput_UnlimitedAndUnset(t *testing.T) {
	out := readFixture(t, "sacctmgr_qos_unlimited.txt")
	q, err := parseQOSOutput(out, "nolimits")
	if err != nil {
		t.Fatalf("parseQOSOutput: %v", err)
	}
	if q.MaxJobsPerUser != datasource.LimitUnset {
		t.Errorf("MaxJobsPerUser = %d, want LimitUnset", q.MaxJobsPerUser)
	}
	if q.MaxSubmitPerUser != datasource.LimitUnlimited {
		t.Errorf("MaxSubmitPerUser = %d, want LimitUnlimited", q.MaxSubmitPerUser)
	}
	if q.MaxWallSet {
		t.Errorf("MaxWallSet = true, want false for UNLIMITED")
	}
}

func TestParseAssocOutput(t *testing.T) {
	out := readFixture(t, "sacctmgr_assoc.txt")
	a, err := parseAssocOutput(out, "myacct", "jdoe", "compute")
	if err != nil {
		t.Fatalf("parseAssocOutput: %v", err)
	}
	if a.MaxJobs != 4 || a.MaxSubmit != 8 {
		t.Errorf("MaxJobs/MaxSubmit = %d/%d, want 4/8", a.MaxJobs, a.MaxSubmit)
	}
}

func TestParseAssocOutput_NotFound(t *testing.T) {
	if _, err := parseAssocOutput([]byte(""), "acct", "user", "part"); err == nil {
		t.Fatal("expected an error for no matching association, got nil")
	}
}

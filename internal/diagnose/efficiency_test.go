package diagnose

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

func TestEfficiencyCheck_Applies(t *testing.T) {
	if !efficiencyApplies(&datasource.Job{State: "COMPLETED"}) {
		t.Error("Applies(COMPLETED) = false, want true")
	}
	if efficiencyApplies(&datasource.Job{State: "FAILED"}) {
		t.Error("Applies(FAILED) = true, want false")
	}
}

func TestRunEfficiencyCheck_Oversized(t *testing.T) {
	job := &datasource.Job{
		ReqMemBytes: 8 * (1 << 30),
		MaxRSSBytes: 100 * (1 << 20), // ~1.2% of requested
		TimeLimit:   2 * time.Hour,
		Elapsed:     5 * time.Minute, // ~4% of requested
	}
	d, err := runEfficiencyCheck(context.Background(), &fakeDataSource{}, job)
	if err != nil {
		t.Fatalf("runEfficiencyCheck: %v", err)
	}
	if d == nil {
		t.Fatal("expected a diagnosis for a dramatically oversized request, got nil")
	}
	if !strings.Contains(d.Why, "memory") || !strings.Contains(d.Why, "time") {
		t.Errorf("Why = %q, want it to flag both memory and time", d.Why)
	}
}

func TestRunEfficiencyCheck_ReasonableUsage_Defers(t *testing.T) {
	job := &datasource.Job{
		ReqMemBytes: 4 * (1 << 30),
		MaxRSSBytes: 3 * (1 << 30), // 75%
		TimeLimit:   time.Hour,
		Elapsed:     50 * time.Minute, // ~83%
	}
	d, err := runEfficiencyCheck(context.Background(), &fakeDataSource{}, job)
	if err != nil {
		t.Fatalf("runEfficiencyCheck: %v", err)
	}
	if d != nil {
		t.Errorf("expected nil for reasonable resource usage, got %+v", d)
	}
}

func TestRunEfficiencyCheck_NoUsageData_Defers(t *testing.T) {
	job := &datasource.Job{} // no ReqMemBytes/MaxRSSBytes/TimeLimit/Elapsed recorded
	d, err := runEfficiencyCheck(context.Background(), &fakeDataSource{}, job)
	if err != nil {
		t.Fatalf("runEfficiencyCheck: %v", err)
	}
	if d != nil {
		t.Errorf("expected nil when there's no usage data to compare, got %+v", d)
	}
}

func TestRunEfficiencyCheck_OnlyMemoryOversized(t *testing.T) {
	job := &datasource.Job{
		ReqMemBytes: 8 * (1 << 30),
		MaxRSSBytes: 100 * (1 << 20),
		TimeLimit:   time.Hour,
		Elapsed:     50 * time.Minute,
	}
	d, err := runEfficiencyCheck(context.Background(), &fakeDataSource{}, job)
	if err != nil {
		t.Fatalf("runEfficiencyCheck: %v", err)
	}
	if d == nil {
		t.Fatal("expected a diagnosis when only memory is oversized, got nil")
	}
	if strings.Contains(d.Why, "time:") {
		t.Errorf("Why = %q, should not flag time when its usage was reasonable", d.Why)
	}
}

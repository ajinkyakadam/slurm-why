package cli

import (
	"testing"
	"time"
)

func TestParseSlurmMemory(t *testing.T) {
	tests := []struct {
		in   string
		want int64
	}{
		{"", 0},
		{"N/A", 0},
		{"4Gn", 4 * (1 << 30)},
		{"512Mc", 512 * (1 << 20)},
		{"1024K", 1024 * (1 << 10)},
		{"2T", 2 * (1 << 40)},
	}
	for _, tt := range tests {
		got, err := parseSlurmMemory(tt.in)
		if err != nil {
			t.Errorf("parseSlurmMemory(%q) returned error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseSlurmMemory(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestParseSlurmMemory_Invalid(t *testing.T) {
	if _, err := parseSlurmMemory("garbage"); err == nil {
		t.Error("expected an error for an unparseable memory string, got nil")
	}
}

func TestParseSlurmDuration(t *testing.T) {
	tests := []struct {
		in      string
		want    time.Duration
		wantSet bool
	}{
		{"UNLIMITED", 0, false},
		{"INFINITE", 0, false},
		{"", 0, false},
		{"00:12:34", 12*time.Minute + 34*time.Second, true},
		{"1-00:00:00", 24 * time.Hour, true},
		{"05:00", 5 * time.Minute, true},
	}
	for _, tt := range tests {
		got, set, err := parseSlurmDuration(tt.in)
		if err != nil {
			t.Errorf("parseSlurmDuration(%q) returned error: %v", tt.in, err)
			continue
		}
		if set != tt.wantSet || got != tt.want {
			t.Errorf("parseSlurmDuration(%q) = (%v,%v), want (%v,%v)", tt.in, got, set, tt.want, tt.wantSet)
		}
	}
}

func TestParseTRES(t *testing.T) {
	got := parseTRES("cpu=4,mem=4G,node=1,gres/gpu=2")
	want := map[string]string{"cpu": "4", "mem": "4G", "node": "1", "gres/gpu": "2"}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("parseTRES()[%q] = %q, want %q", k, got[k], v)
		}
	}
}

func TestParsePipeRows_TrailingDelimiter(t *testing.T) {
	rows := parsePipeRows([]byte("a|b|\n"), []string{"first", "second"})
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	if rows[0]["first"] != "a" || rows[0]["second"] != "b" {
		t.Errorf("row = %+v, want first=a second=b", rows[0])
	}
}

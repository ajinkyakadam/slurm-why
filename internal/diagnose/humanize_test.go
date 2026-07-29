package diagnose

import "testing"

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		in   int64
		want string
	}{
		{512, "512B"},
		{2 * (1 << 10), "2.0K"},
		{int64(4.5 * (1 << 30)), "4.5G"},
		{2 * (1 << 40), "2.0T"},
	}
	for _, tt := range tests {
		if got := formatBytes(tt.in); got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

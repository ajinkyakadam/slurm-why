package diagnose

import "fmt"

// formatBytes renders a byte count as a Slurm-style --mem value (e.g.
// "4.0G"), using the largest unit that keeps the value >= 1.
func formatBytes(n int64) string {
	units := []struct {
		suffix string
		size   int64
	}{
		{"T", 1 << 40},
		{"G", 1 << 30},
		{"M", 1 << 20},
		{"K", 1 << 10},
	}
	for _, u := range units {
		if n >= u.size {
			return fmt.Sprintf("%.1f%s", float64(n)/float64(u.size), u.suffix)
		}
	}
	return fmt.Sprintf("%dB", n)
}

// Package format renders a diagnosis as Diagnosis -> Why -> Fix. It has no
// dependency on Slurm-specific types, so the same rendering can be reused
// by future sibling tools (kueue-why, spack-why) that share this catalog
// format.
package format

import "fmt"

// Diagnosis is a single plain-English explanation for why a job did what
// it did, and what to do about it.
type Diagnosis struct {
	Title string
	Why   string
	Fix   string
}

// Render formats a Diagnosis as human-readable text.
func Render(d Diagnosis) string {
	return fmt.Sprintf("%s\n\nWhy: %s\n\nFix: %s\n", d.Title, d.Why, d.Fix)
}

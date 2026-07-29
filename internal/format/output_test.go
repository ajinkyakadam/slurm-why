package format

import (
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	d := Diagnosis{
		Title: "Job was killed for using too much memory (OOM)",
		Why:   "it used more memory than requested",
		Fix:   "increase --mem",
	}
	got := Render(d)

	for _, want := range []string{d.Title, d.Why, d.Fix, "Why:", "Fix:"} {
		if !strings.Contains(got, want) {
			t.Errorf("Render() = %q, want it to contain %q", got, want)
		}
	}
}

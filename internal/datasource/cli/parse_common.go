package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

// memoryPattern matches a Slurm memory value: a number, an optional
// K/M/G/T/P unit suffix, and an optional trailing n/c (per-node / per-cpu
// indicator used on ReqMem, never on MaxRSS). Case-insensitive.
var memoryPattern = regexp.MustCompile(`(?i)^([0-9.]+)([KMGTP]?)[NC]?$`)

var memoryUnits = map[string]int64{
	"":  1, // sacct/scontrol only omit the suffix for values already in bytes
	"K": 1 << 10,
	"M": 1 << 20,
	"G": 1 << 30,
	"T": 1 << 40,
	"P": 1 << 50,
}

// parseSlurmMemory converts a Slurm memory string (e.g. "4Gn", "3202396K",
// "512M") to bytes. An empty string returns 0 with no error, since sacct
// commonly leaves MaxRSS blank for steps that never reported usage.
func parseSlurmMemory(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "N/A") {
		return 0, nil
	}
	m := memoryPattern.FindStringSubmatch(s)
	if m == nil {
		return 0, fmt.Errorf("parseSlurmMemory: unrecognized value %q", s)
	}
	value, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, fmt.Errorf("parseSlurmMemory: %q: %w", s, err)
	}
	unit := strings.ToUpper(m[2])
	return int64(value * float64(memoryUnits[unit])), nil
}

// parseSlurmDuration parses a Slurm elapsed/timelimit string. Slurm uses
// "HH:MM:SS", "MM:SS", or "D-HH:MM:SS" for real durations, and reports
// "UNLIMITED", "INVALID", "N/A", or "Partition_Limit" for the sentinel
// cases — the second return value is false for all of those.
func parseSlurmDuration(s string) (time.Duration, bool, error) {
	s = strings.TrimSpace(s)
	switch {
	case s == "", strings.EqualFold(s, "UNLIMITED"), strings.EqualFold(s, "INFINITE"),
		strings.EqualFold(s, "INVALID"), strings.EqualFold(s, "N/A"), strings.EqualFold(s, "NONE"),
		strings.EqualFold(s, "Partition_Limit"):
		return 0, false, nil
	}

	days := 0
	rest := s
	if idx := strings.Index(s, "-"); idx >= 0 {
		d, err := strconv.Atoi(s[:idx])
		if err != nil {
			return 0, false, fmt.Errorf("parseSlurmDuration: %q: bad day component: %w", s, err)
		}
		days = d
		rest = s[idx+1:]
	}

	parts := strings.Split(rest, ":")
	var hours, minutes, seconds int
	var err error
	switch len(parts) {
	case 3:
		hours, err = strconv.Atoi(parts[0])
		if err == nil {
			minutes, err = strconv.Atoi(parts[1])
		}
		if err == nil {
			seconds, err = strconv.Atoi(parts[2])
		}
	case 2:
		minutes, err = strconv.Atoi(parts[0])
		if err == nil {
			seconds, err = strconv.Atoi(parts[1])
		}
	default:
		return 0, false, fmt.Errorf("parseSlurmDuration: %q: unrecognized format", s)
	}
	if err != nil {
		return 0, false, fmt.Errorf("parseSlurmDuration: %q: %w", s, err)
	}

	total := time.Duration(days)*24*time.Hour +
		time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second
	return total, true, nil
}

// slurmTimeLayout is the timestamp format scontrol uses for StartTime/
// EndTime fields, e.g. "2026-07-30T05:35:02" (no timezone -- always the
// scheduler's local time).
const slurmTimeLayout = "2006-01-02T15:04:05"

// parseSlurmTime parses a scontrol timestamp field, returning the zero
// time (no error) for Slurm's "not applicable yet" sentinels.
func parseSlurmTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "Unknown") || strings.EqualFold(s, "N/A") || strings.EqualFold(s, "NONE") {
		return time.Time{}, nil
	}
	t, err := time.Parse(slurmTimeLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parseSlurmTime: %q: %w", s, err)
	}
	return t, nil
}

// parseTRES splits a Slurm TRES string ("cpu=4,mem=4G,node=1,gres/gpu=2")
// into a flat key->value map. Keys keep any "gres/" prefix as-is.
func parseTRES(s string) map[string]string {
	out := map[string]string{}
	s = strings.TrimSpace(s)
	if s == "" {
		return out
	}
	for _, pair := range strings.Split(s, ",") {
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return out
}

// parsePipeRows splits parsable2/noheader output into rows, zipping each
// row's "|"-delimited columns with fields by position. A trailing empty
// column (from a trailing delimiter) is tolerated and dropped.
func parsePipeRows(output []byte, fields []string) []map[string]string {
	var rows []map[string]string
	for _, line := range strings.Split(string(output), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		cols := strings.Split(line, "|")
		if len(cols) > len(fields) {
			cols = cols[:len(fields)]
		}
		row := make(map[string]string, len(fields))
		for i, f := range fields {
			if i < len(cols) {
				row[f] = cols[i]
			} else {
				row[f] = ""
			}
		}
		rows = append(rows, row)
	}
	return rows
}

// keyValuePattern extracts "Key=Value" tokens from scontrol's whitespace
// (spaces and newlines) separated block output. Values containing spaces
// (e.g. Command=...) will only capture up to the first space, which is
// acceptable since none of the fields this tool reads contain spaces.
var keyValuePattern = regexp.MustCompile(`([A-Za-z][A-Za-z0-9_:/]*)=(\S*)`)

// parseKeyValueBlob extracts all Key=Value tokens from a scontrol
// "show job"/"show partition" text block into a flat map.
func parseKeyValueBlob(output []byte) map[string]string {
	out := map[string]string{}
	for _, m := range keyValuePattern.FindAllStringSubmatch(string(output), -1) {
		out[m[1]] = m[2]
	}
	return out
}

// parseSlurmIntLimit parses an integer ceiling field (MaxNodes,
// MaxJobsPerUser, etc.) that Slurm may instead report as UNLIMITED/INFINITE
// or leave blank/N/A, mapping those to the datasource sentinels.
func parseSlurmIntLimit(s string) int {
	s = strings.TrimSpace(s)
	switch {
	case s == "", strings.EqualFold(s, "N/A"), strings.EqualFold(s, "NONE"):
		return datasource.LimitUnset
	case strings.EqualFold(s, "UNLIMITED"), strings.EqualFold(s, "INFINITE"), s == "-1":
		return datasource.LimitUnlimited
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return datasource.LimitUnset
	}
	return n
}

// parseSlurmMemLimit parses a memory ceiling field (MaxMemPerNode, etc.)
// that Slurm may report as UNLIMITED or leave blank, mapping those to the
// datasource sentinels. Bare numbers (no unit suffix) are MB, matching
// scontrol's convention for these specific fields.
func parseSlurmMemLimit(s string) int64 {
	s = strings.TrimSpace(s)
	switch {
	case s == "", strings.EqualFold(s, "N/A"), strings.EqualFold(s, "NONE"):
		return datasource.LimitUnset
	case strings.EqualFold(s, "UNLIMITED"), strings.EqualFold(s, "INFINITE"), s == "-1":
		return datasource.LimitUnlimited
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n * (1 << 20)
	}
	if bytes, err := parseSlurmMemory(s); err == nil {
		return bytes
	}
	return datasource.LimitUnset
}

// atoiOr0 parses s as an int, returning 0 for empty/unparseable input
// rather than an error — used for fields where "missing" and "zero" are
// both acceptable outcomes for a best-effort diagnosis.
func atoiOr0(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return n
}

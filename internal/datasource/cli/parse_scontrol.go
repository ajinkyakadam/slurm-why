package cli

import (
	"context"
	"fmt"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

// GetPartitionLimits fetches the ceilings a partition can ever grant a job
// via scontrol show partition. Partition configuration is public in Slurm
// (users need it to pick a partition), so this requires no special access.
func (c *Client) GetPartitionLimits(ctx context.Context, partition string) (*datasource.PartitionLimits, error) {
	out, err := c.run(ctx, "scontrol", "show", "partition", partition)
	if err != nil {
		return nil, fmt.Errorf("scontrol show partition: %w", err)
	}
	return parsePartitionOutput(out, partition)
}

// GetDependency fetches a job's raw --dependency spec via scontrol show
// job. Not available from sacct (no such format field), which is why this
// is a separate lookup made only when a dependency check actually needs
// it, rather than joined into every GetJob call.
func (c *Client) GetDependency(ctx context.Context, jobID string) (*datasource.Dependency, error) {
	out, err := c.run(ctx, "scontrol", "show", "job", jobID)
	if err != nil {
		return nil, fmt.Errorf("scontrol show job: %w", err)
	}
	return parseDependencyOutput(out, jobID)
}

func parseDependencyOutput(out []byte, jobID string) (*datasource.Dependency, error) {
	kv := parseKeyValueBlob(out)
	if len(kv) == 0 {
		return nil, fmt.Errorf("scontrol show job: no job %q found", jobID)
	}
	raw := kv["Dependency"]
	if raw == "(null)" {
		raw = ""
	}
	return &datasource.Dependency{Raw: raw}, nil
}

func parsePartitionOutput(out []byte, partition string) (*datasource.PartitionLimits, error) {
	kv := parseKeyValueBlob(out)
	if len(kv) == 0 {
		return nil, fmt.Errorf("scontrol show partition: no partition named %q found", partition)
	}

	maxTime, maxTimeSet, err := parseSlurmDuration(kv["MaxTime"])
	if err != nil {
		return nil, fmt.Errorf("scontrol show partition: MaxTime: %w", err)
	}

	return &datasource.PartitionLimits{
		Name:          kv["PartitionName"],
		MaxTimeLimit:  maxTime,
		MaxTimeSet:    maxTimeSet,
		MaxNodes:      parseSlurmIntLimit(kv["MaxNodes"]),
		MaxMemPerNode: parseSlurmMemLimit(kv["MaxMemPerNode"]),
	}, nil
}

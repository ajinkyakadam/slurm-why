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

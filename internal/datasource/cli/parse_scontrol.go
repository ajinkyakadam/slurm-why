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

// GetReservation looks up the reservation a job explicitly requested via
// --reservation=<name>. It first reads the job's own Reservation= field
// (scontrol show job) -- a job blocked by *someone else's* unrelated
// reservation has no such field, only a free-text annotation on its
// pending reason (see PendingReason.Detail), and there's no reliable way
// to identify which reservation that is without guessing, so this only
// handles the direct case.
func (c *Client) GetReservation(ctx context.Context, jobID string) (*datasource.ReservationInfo, error) {
	jobOut, err := c.run(ctx, "scontrol", "show", "job", jobID)
	if err != nil {
		return nil, fmt.Errorf("scontrol show job: %w", err)
	}
	kv := parseKeyValueBlob(jobOut)
	if len(kv) == 0 {
		return nil, fmt.Errorf("scontrol show job: no job %q found", jobID)
	}
	name := kv["Reservation"]
	if name == "" || name == "(null)" {
		return nil, nil
	}

	resOut, err := c.run(ctx, "scontrol", "show", "reservation", name)
	if err != nil {
		return nil, fmt.Errorf("scontrol show reservation: %w", err)
	}
	return parseReservationOutput(resOut, name)
}

func parseReservationOutput(out []byte, name string) (*datasource.ReservationInfo, error) {
	kv := parseKeyValueBlob(out)
	if len(kv) == 0 {
		return nil, fmt.Errorf("scontrol show reservation: no reservation named %q found", name)
	}

	start, err := parseSlurmTime(kv["StartTime"])
	if err != nil {
		return nil, fmt.Errorf("scontrol show reservation: StartTime: %w", err)
	}
	end, err := parseSlurmTime(kv["EndTime"])
	if err != nil {
		return nil, fmt.Errorf("scontrol show reservation: EndTime: %w", err)
	}

	return &datasource.ReservationInfo{
		Name:      kv["ReservationName"],
		StartTime: start,
		EndTime:   end,
		Nodes:     kv["Nodes"],
		State:     kv["State"],
	}, nil
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

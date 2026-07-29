package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

var qosFields = []string{"Name", "MaxJobsPerUser", "MaxSubmitJobsPerUser", "MaxWall"}

// GetQOSLimits fetches per-user ceilings for a QOS via sacctmgr show qos.
// Viewing QOS definitions is read-visible to any user even on sites that
// restrict write access to sacctmgr — no elevated privileges required.
func (c *Client) GetQOSLimits(ctx context.Context, qos string) (*datasource.QOSLimits, error) {
	out, err := c.run(ctx, "sacctmgr", "show", "qos", qos,
		"--parsable2", "--noheader", "--format="+strings.Join(qosFields, ","),
	)
	if err != nil {
		return nil, fmt.Errorf("sacctmgr show qos: %w", err)
	}
	return parseQOSOutput(out, qos)
}

func parseQOSOutput(out []byte, qos string) (*datasource.QOSLimits, error) {
	rows := parsePipeRows(out, qosFields)
	if len(rows) == 0 {
		return nil, fmt.Errorf("sacctmgr show qos: no QOS named %q found", qos)
	}
	row := rows[0]

	maxWall, maxWallSet, err := parseSlurmDuration(row["MaxWall"])
	if err != nil {
		return nil, fmt.Errorf("sacctmgr show qos: MaxWall: %w", err)
	}

	return &datasource.QOSLimits{
		Name:             row["Name"],
		MaxJobsPerUser:   parseSlurmIntLimit(row["MaxJobsPerUser"]),
		MaxSubmitPerUser: parseSlurmIntLimit(row["MaxSubmitJobsPerUser"]),
		MaxWall:          maxWall,
		MaxWallSet:       maxWallSet,
	}, nil
}

var assocFields = []string{"Account", "User", "Partition", "MaxJobs", "MaxSubmitJobs"}

// GetAssocLimits fetches the ceilings on a specific account/user/partition
// association via sacctmgr show assoc. Same read-only, no-elevated-access
// note as GetQOSLimits applies.
func (c *Client) GetAssocLimits(ctx context.Context, account, user, partition string) (*datasource.AssocLimits, error) {
	out, err := c.run(ctx, "sacctmgr", "show", "assoc",
		"account="+account, "user="+user, "partition="+partition,
		"--parsable2", "--noheader", "--format="+strings.Join(assocFields, ","),
	)
	if err != nil {
		return nil, fmt.Errorf("sacctmgr show assoc: %w", err)
	}
	return parseAssocOutput(out, account, user, partition)
}

func parseAssocOutput(out []byte, account, user, partition string) (*datasource.AssocLimits, error) {
	rows := parsePipeRows(out, assocFields)
	if len(rows) == 0 {
		return nil, fmt.Errorf("sacctmgr show assoc: no association found for account=%s user=%s partition=%s", account, user, partition)
	}
	row := rows[0]

	return &datasource.AssocLimits{
		Account:   row["Account"],
		User:      row["User"],
		Partition: row["Partition"],
		MaxJobs:   parseSlurmIntLimit(row["MaxJobs"]),
		MaxSubmit: parseSlurmIntLimit(row["MaxSubmitJobs"]),
	}, nil
}

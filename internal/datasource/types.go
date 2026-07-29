// Package datasource defines the backend-agnostic view of Slurm job/limit
// data that the diagnosis checks consume. Two backends implement
// JobDataSource: a REST client (slurmrestd, when a site has it configured)
// and a CLI client (sacct/squeue/scontrol/sacctmgr, always available on a
// login node). Resolve picks between them at startup.
package datasource

import (
	"context"
	"time"
)

// Sentinel values used for limit fields where Slurm distinguishes "not set"
// from "explicitly unlimited" from "a real numeric ceiling".
const (
	// LimitUnset means the field could not be determined (e.g. missing account
	// association, or the site doesn't set a limit for this QOS/partition).
	LimitUnset = 0
	// LimitUnlimited means Slurm reported the ceiling as UNLIMITED/-1.
	LimitUnlimited = -1
)

// Job is the subset of sacct/scontrol job data the v0.1 checks need.
type Job struct {
	ID         string
	State      string // e.g. RUNNING, COMPLETED, OUT_OF_MEMORY, PENDING, TIMEOUT
	ExitCode   int
	ExitSignal int
	Account    string
	Partition  string
	QOS        string
	User       string

	ReqMemBytes int64 // requested memory, in bytes (0 if unknown)
	MaxRSSBytes int64 // peak observed RSS across all job steps, in bytes (0 if unknown)

	ReqCPUs  int
	ReqNodes int
	ReqGPUs  int

	Elapsed   time.Duration
	TimeLimit time.Duration
}

// PendingReason is the decoded NODELIST(REASON) info for a job in state PD.
type PendingReason struct {
	Code string // raw Slurm reason code, e.g. "Priority", "Resources", "QOSMaxJobsPerUserLimit"
}

// PartitionLimits are the ceilings a partition can ever grant a job.
type PartitionLimits struct {
	Name string

	MaxTimeLimit  time.Duration // 0 if unknown; check MaxTimeSet before treating 0 as "no limit"
	MaxTimeSet    bool
	MaxNodes      int   // LimitUnlimited if UNLIMITED, LimitUnset if unknown
	MaxMemPerNode int64 // bytes; LimitUnlimited if UNLIMITED, LimitUnset if unknown
}

// QOSLimits are the per-user ceilings a QOS enforces.
type QOSLimits struct {
	Name string

	MaxJobsPerUser   int // LimitUnlimited or LimitUnset
	MaxSubmitPerUser int
	MaxWall          time.Duration
	MaxWallSet       bool
}

// AssocLimits are the ceilings on a specific account/user/partition association.
type AssocLimits struct {
	Account   string
	User      string
	Partition string

	MaxJobs   int // LimitUnlimited or LimitUnset
	MaxSubmit int
}

// JobDataSource is implemented by both the REST and CLI backends. All
// methods are read-only — none of the diagnosis checks need write access to
// Slurm, only visibility into a job's own state and the cluster's published
// limits.
type JobDataSource interface {
	GetJob(ctx context.Context, jobID string) (*Job, error)
	GetPendingReason(ctx context.Context, jobID string) (*PendingReason, error)
	GetPartitionLimits(ctx context.Context, partition string) (*PartitionLimits, error)
	GetQOSLimits(ctx context.Context, qos string) (*QOSLimits, error)
	GetAssocLimits(ctx context.Context, account, user, partition string) (*AssocLimits, error)

	// Backend identifies which implementation is in use ("cli" or "rest"),
	// surfaced to the user via -v/--verbose.
	Backend() string
}

package diagnose

import (
	"context"
	"errors"
	"fmt"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

// errBoom is a sentinel error tests use to verify error propagation.
var errBoom = errors.New("boom")

// fakeDataSource is a test double for datasource.JobDataSource, letting
// tests stub out just the lookups a given check needs without shelling
// out to real Slurm commands.
type fakeDataSource struct {
	pendingReason   *datasource.PendingReason
	pendingErr      error
	partitionLimits map[string]*datasource.PartitionLimits
	qosLimits       map[string]*datasource.QOSLimits
	dependency      *datasource.Dependency
	dependencyErr   error
	pendingJobIDs   []string
	pendingJobsErr  error
}

func (f *fakeDataSource) GetJob(ctx context.Context, jobID string) (*datasource.Job, error) {
	return nil, fmt.Errorf("fakeDataSource.GetJob not used in these tests")
}

func (f *fakeDataSource) GetPendingReason(ctx context.Context, jobID string) (*datasource.PendingReason, error) {
	return f.pendingReason, f.pendingErr
}

func (f *fakeDataSource) GetPartitionLimits(ctx context.Context, partition string) (*datasource.PartitionLimits, error) {
	if p, ok := f.partitionLimits[partition]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("fakeDataSource: no partition %q configured", partition)
}

func (f *fakeDataSource) GetQOSLimits(ctx context.Context, qos string) (*datasource.QOSLimits, error) {
	if q, ok := f.qosLimits[qos]; ok {
		return q, nil
	}
	return nil, fmt.Errorf("fakeDataSource: no QOS %q configured", qos)
}

func (f *fakeDataSource) GetAssocLimits(ctx context.Context, account, user, partition string) (*datasource.AssocLimits, error) {
	return nil, fmt.Errorf("fakeDataSource.GetAssocLimits not used in these tests")
}

func (f *fakeDataSource) GetDependency(ctx context.Context, jobID string) (*datasource.Dependency, error) {
	return f.dependency, f.dependencyErr
}

func (f *fakeDataSource) GetPendingJobIDs(ctx context.Context) ([]string, error) {
	return f.pendingJobIDs, f.pendingJobsErr
}

func (f *fakeDataSource) Backend() string { return "fake" }

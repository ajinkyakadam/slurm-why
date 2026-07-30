package diagnose

import (
	"context"
	"fmt"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
	"github.com/ajinkyakadam/slurm-why/internal/format"
)

func init() {
	// Default Priority (0) -- must run before "pending" (pendingRunLast)
	// so its richer, live-looked-up explanation wins over the static
	// "Reservation" entry in pendingReasons for jobs this applies to.
	Registry = append(Registry, Check{ID: "reservation", Applies: reservationApplies, Run: runReservationCheck})
}

func reservationApplies(j *datasource.Job) bool {
	return j.State == "PENDING" || j.State == "PD"
}

const reservationTimeFormat = "2006-01-02 15:04"

func runReservationCheck(ctx context.Context, ds datasource.JobDataSource, job *datasource.Job) (*format.Diagnosis, error) {
	res, err := ds.GetReservation(ctx, job.ID)
	if err != nil {
		return nil, err
	}
	if res == nil {
		// Job doesn't reference a reservation directly -- defer to the
		// generic pending check (which still explains a job blocked by
		// someone *else's* reservation, via the Detail-suffix on
		// ReqNodeNotAvail).
		return nil, nil
	}

	window := fmt.Sprintf("%s to %s", res.StartTime.Format(reservationTimeFormat), res.EndTime.Format(reservationTimeFormat))

	if res.State == "ACTIVE" {
		return &format.Diagnosis{
			Title: "Job is waiting within its own active reservation",
			Why: fmt.Sprintf(
				"This job requested reservation %q, which is already active (%s) but the job still hasn't started -- likely waiting on capacity within the reservation's own node/CPU allocation, not the reservation window itself.",
				res.Name, window),
			Fix: fmt.Sprintf("Run `scontrol show reservation %s` to check its remaining capacity, or ask your site admin if it looks stuck despite the reservation being active.", res.Name),
		}, nil
	}

	return &format.Diagnosis{
		Title: "Job is waiting for its reservation to start",
		Why: fmt.Sprintf(
			"This job requested reservation %q, scheduled %s, which hasn't started yet.",
			res.Name, window),
		Fix: fmt.Sprintf("This is normal -- no action needed. It'll become eligible once the reservation starts at %s.", res.StartTime.Format(reservationTimeFormat)),
	}, nil
}

package diagnose

import (
	"context"
	"testing"
	"time"

	"github.com/ajinkyakadam/slurm-why/internal/datasource"
)

func TestReservationCheck_Applies(t *testing.T) {
	if !reservationApplies(&datasource.Job{State: "PENDING"}) {
		t.Error("Applies(PENDING) = false, want true")
	}
	if reservationApplies(&datasource.Job{State: "RUNNING"}) {
		t.Error("Applies(RUNNING) = true, want false")
	}
}

func TestRunReservationCheck_NotStarted(t *testing.T) {
	job := &datasource.Job{ID: "44"}
	ds := &fakeDataSource{reservation: &datasource.ReservationInfo{
		Name:      "maint5",
		StartTime: time.Date(2026, 7, 30, 8, 0, 0, 0, time.UTC),
		EndTime:   time.Date(2026, 7, 30, 9, 0, 0, 0, time.UTC),
		Nodes:     "slurmtest",
		State:     "INACTIVE",
	}}

	d, err := runReservationCheck(context.Background(), ds, job)
	if err != nil {
		t.Fatalf("runReservationCheck: %v", err)
	}
	if d == nil {
		t.Fatal("expected a diagnosis, got nil")
	}
	if d.Title != "Job is waiting for its reservation to start" {
		t.Errorf("Title = %q", d.Title)
	}
}

func TestRunReservationCheck_ActiveButNotStarted(t *testing.T) {
	job := &datasource.Job{ID: "43"}
	ds := &fakeDataSource{reservation: &datasource.ReservationInfo{
		Name:  "maint4",
		State: "ACTIVE",
	}}

	d, err := runReservationCheck(context.Background(), ds, job)
	if err != nil {
		t.Fatalf("runReservationCheck: %v", err)
	}
	if d.Title != "Job is waiting within its own active reservation" {
		t.Errorf("Title = %q", d.Title)
	}
}

func TestRunReservationCheck_NoReservationDefers(t *testing.T) {
	job := &datasource.Job{ID: "36"}
	ds := &fakeDataSource{reservation: nil}

	d, err := runReservationCheck(context.Background(), ds, job)
	if err != nil {
		t.Fatalf("runReservationCheck: %v", err)
	}
	if d != nil {
		t.Errorf("expected nil (defer to pending check) when job has no reservation, got %+v", d)
	}
}

func TestRunReservationCheck_PropagatesError(t *testing.T) {
	job := &datasource.Job{ID: "44"}
	ds := &fakeDataSource{reservationErr: errBoom}

	if _, err := runReservationCheck(context.Background(), ds, job); err == nil {
		t.Error("expected an error to propagate from GetReservation, got nil")
	}
}

// TestReservationCheck_BeatsGenericPendingInRegistry guards the ordering
// this check depends on: it must run before "pending" (Priority
// pendingRunLast) so a job with a reservation gets the live-looked-up
// window, not just the static "Reservation" text in pendingReasons.
func TestReservationCheck_BeatsGenericPendingInRegistry(t *testing.T) {
	job := &datasource.Job{ID: "44", State: "PENDING"}
	ds := &fakeDataSource{
		reservation: &datasource.ReservationInfo{
			Name:      "maint5",
			StartTime: time.Date(2026, 7, 30, 8, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 7, 30, 9, 0, 0, 0, time.UTC),
			State:     "INACTIVE",
		},
		pendingReason: &datasource.PendingReason{Code: "Reservation"},
	}

	got, checkID, err := DiagnoseWithCheckID(context.Background(), ds, job, Registry)
	if err != nil {
		t.Fatalf("DiagnoseWithCheckID: %v", err)
	}
	if checkID != "reservation" {
		t.Errorf("checkID = %q, want %q (reservation check should win over the generic pending catch-all)", checkID, "reservation")
	}
	if got.Title != "Job is waiting for its reservation to start" {
		t.Errorf("Title = %q", got.Title)
	}
}

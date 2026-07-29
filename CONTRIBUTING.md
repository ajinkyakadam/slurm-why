# Contributing to slurm-why

## Prerequisites

- Go 1.23+
- git

## Building

```
git clone https://github.com/ajinkyakadam/slurm-why.git
cd slurm-why
go build -o slurm-why ./cmd/slurm-why
./slurm-why version
```

## Running the test suite

```
go build ./...
go vet ./...
go test ./... -race -cover
```

All tests are fixture-based (see `internal/datasource/cli/testdata/`) and never invoke real `sacct`/`squeue`/`scontrol`/`sacctmgr` binaries — they substitute canned output through an injectable `Runner`. The full suite runs identically whether or not Slurm is installed on your machine.

To run one package's tests verbosely:

```
go test ./internal/diagnose/... -v
```

## Trying the CLI locally without a real Slurm cluster

The CLI backend shells out to whatever `sacct`/`squeue`/`scontrol`/`sacctmgr` resolve to on `$PATH`. On a laptop without Slurm installed, running the built binary fails immediately with a clear "executable file not found" error — there's nothing to talk to.

To manually exercise the *full CLI* (not just the unit tests) against canned output, shadow those commands with tiny scripts that `cat` a fixture file, and put them ahead of the (nonexistent) real ones on `$PATH`. For example, to see the OOM diagnosis:

```bash
mkdir -p /tmp/fake-slurm
cat > /tmp/fake-slurm/sacct <<'EOF'
#!/bin/sh
cat /path/to/slurm-why/internal/datasource/cli/testdata/sacct_oom.txt
EOF
chmod +x /tmp/fake-slurm/sacct

PATH=/tmp/fake-slurm:$PATH ./slurm-why 123
```

(substitute your actual clone path). This prints the OOM diagnosis computed from the exact fixture used in `internal/datasource/cli/parse_sacct_test.go` — the job's `MaxRSS` genuinely comes in below its `ReqMem`, so this also demonstrates the sacct-sampling-vs-cgroup explanation.

To try the pending-reason diagnosis, `sacct` needs to report a `PENDING` job (there's no ready-made fixture for this — the existing `sacct_*.txt` fixtures are both terminal-state examples — so a one-line inline stub is easier than a file), and `squeue` needs the reason code:

```bash
cat > /tmp/fake-slurm/sacct <<'EOF'
#!/bin/sh
echo "555|PENDING|0:0|myacct|compute|normal|jdoe|4Gn||00:00:00|01:00:00|cpu=2,mem=4G,node=1"
EOF
cat > /tmp/fake-slurm/squeue <<'EOF'
#!/bin/sh
cat /path/to/slurm-why/internal/datasource/cli/testdata/squeue_pending_priority.txt
EOF
chmod +x /tmp/fake-slurm/sacct /tmp/fake-slurm/squeue

PATH=/tmp/fake-slurm:$PATH ./slurm-why 555
```

Swap in any of the other `squeue_pending_*.txt` fixtures (or the `scontrol_partition_*.txt`/`sacctmgr_*.txt` ones, plus a fake `scontrol`/`sacctmgr`, for the impossible-request check) to see the other diagnoses render.

This is a manual smoke-testing technique, not a supported CLI feature — it's useful for a quick sanity check while developing, but it doesn't exercise real Slurm scheduling behavior (queue state changes, actual resource enforcement). Full integration testing against a live cluster is tracked as a separate effort.

## Adding a new diagnosis check

Each check lives in its own file in `internal/diagnose/` and self-registers into the shared registry via an `init()` function — see `oom.go`, `pending.go`, or `impossible.go` for the pattern. A check needs:

- `Applies(*datasource.Job) bool` — a cheap, synchronous decision on whether this check might have something to say
- `Run(ctx, datasource.JobDataSource, *datasource.Job) (*format.Diagnosis, error)` — the actual logic; return `nil, nil` (not an error) if, after fetching more data, there's nothing specific to report, so a later, more generic check gets a turn instead

Add fixture-based tests alongside it (see `internal/datasource/cli/testdata/` for the pattern of capturing real command output as a test fixture).

## Code style

- `gofmt` — no unformatted code
- `go vet` clean
- No comments explaining *what* code does — only comment non-obvious *why* (a hidden constraint, a workaround, a subtle invariant)

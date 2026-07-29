# Local test Slurm cluster

A minimal single-node Slurm cluster for manually validating slurm-why's diagnoses against real `slurmctld`/`slurmdbd`/`sacct`/`squeue`/`scontrol` behavior — unit tests use fixtures and never touch real Slurm, so this is the only way to confirm slurm-why's assumptions actually hold against a live cluster.

**Provenance**: built from the official `ubuntu` Docker image, pinned to a specific content digest (not a floating tag), with Slurm installed via Ubuntu's own signed `slurm-wlm` apt package — no third-party Docker images, no compiling Slurm from source.

## Bring it up

```
cd test/cluster
docker compose up -d --build
```

Give it ~15-20 seconds to finish starting `munged`, MariaDB, `slurmdbd`, `slurmctld`, and `slurmd`, then confirm it's healthy:

```
docker exec slurm-why-testcluster sinfo
```

You should see partition `debug*` with node `slurmtest` in state `idle`.

No persistent volumes are used on purpose — every `docker compose up --build` gives a genuinely fresh cluster (empty accounting DB, no leftover job history). Re-initializing MariaDB only takes a few seconds, and reproducibility across runs matters more here than fast restarts.

## Cluster limits (deliberately tight, to make violations easy to trigger)

- 1 node, 2 CPUs, ~3.8GB RAM
- Partition `debug`: `MaxNodes=1`, `MaxMemPerNode=2000` (MB), `MaxTime=01:00:00`

## Example jobs

`jobs/` has one script per scenario:

| Script | Triggers |
|---|---|
| `normal.sh` | Completes successfully — validates the fallback "completed" diagnosis |
| `impossible_nodes.sh` | Requests `--nodes=2` > partition's `MaxNodes=1` → reason `PartitionNodeLimit` |
| `impossible_mem.sh` | Requests `--mem=3000M` > partition's `MaxMemPerNode=2000` → reason `MaxMemPerLimit` |
| `impossible_time.sh` | Requests `--time=02:00:00` > partition's `MaxTime=01:00:00` → reason `PartitionTimeLimit` |

Submit them as `testuser` (the only account configured in the test cluster's accounting DB):

```
docker cp jobs/normal.sh slurm-why-testcluster:/home/testuser/normal.sh
docker exec slurm-why-testcluster chown testuser:testuser /home/testuser/normal.sh
docker exec --user testuser slurm-why-testcluster bash -c "cd /home/testuser && sbatch normal.sh"
```

(repeat per script). Check results with `docker exec slurm-why-testcluster squeue` / `sacct`.

## Running slurm-why against it

The container is `linux/arm64` (or `linux/amd64`, matching your host) — build slurm-why for that target, copy it in, and run it as `testuser` so its `sacct`/`squeue`/etc. calls see `testuser`'s own jobs:

```
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o /tmp/slurm-why-linux-arm64 ../../cmd/slurm-why
docker cp /tmp/slurm-why-linux-arm64 slurm-why-testcluster:/usr/local/bin/slurm-why
docker exec slurm-why-testcluster chmod +x /usr/local/bin/slurm-why

docker exec --user testuser slurm-why-testcluster slurm-why <jobid>
```

(swap `arm64` for `amd64` if your Docker host is x86_64.)

## Known limitation: no OOM testing yet

This cluster doesn't configure cgroup-based memory enforcement (`TaskPlugin=task/affinity`, no `cgroup.conf`), so there's no way to trigger a genuine Slurm-issued `OUT_OF_MEMORY` kill here yet — getting cgroup v2 delegation working reliably inside Docker Desktop's nested VM turned out to be significant additional complexity, deferred rather than rushed. The OOM check itself is validated via the fixture-based unit tests (`internal/datasource/cli/testdata/sacct_oom.txt`) and the manual fixture-shadowing technique in `CONTRIBUTING.md`; this cluster currently covers the pending-reason decoder and impossible-request checks against real Slurm output, plus the completed-job fallback.

## Tear down

```
docker compose down
```

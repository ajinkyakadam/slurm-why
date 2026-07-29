# Local test Slurm cluster

A minimal single-node Slurm cluster for manually validating slurm-why's diagnoses against real `slurmctld`/`slurmdbd`/`sacct`/`squeue`/`scontrol` behavior — unit tests use fixtures and never touch real Slurm, so this is the only way to confirm slurm-why's assumptions actually hold against a live cluster.

**Provenance**: built from the official `ubuntu` Docker image, pinned to a specific content digest (not a floating tag). Slurm itself is compiled from SchedMD's own official GitHub release source tarball (not a distro package — no distro ships anything close to current — and not a third-party image), with the download verified against the exact SHA256 digest GitHub's release API reports for that asset. Currently pinned to **26.05.2** (`Dockerfile` `ARG`s `SLURM_VERSION`/`SLURM_RELEASE_TAG`/`SLURM_SHA256`) — check [SchedMD's releases](https://github.com/SchedMD/slurm/releases) for the current version before bumping.

Not every site runs the latest Slurm, though — this project's parsing is deliberately built to work across a wide version range (see `PLAN.md`). Worth periodically rebuilding this cluster against an older version too, to confirm nothing's been accidentally narrowed to only-latest-Slurm behavior.

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
| `oom.sh` | Requests `--mem=100M` then grows unbounded — killed by the real memory cgroup |

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

## Real cgroup memory enforcement (and what it took)

`oom.sh` triggers a genuine memory-cgroup kill, not a simulated one. Getting there required more than `TaskPlugin=task/cgroup` + `cgroup.conf`:

- **Slurm 26.05's `cgroup/v2` plugin needs `libdbus`** to build at all (the Dockerfile installs `libdbus-1-dev`/`libhwloc-dev`) — without it, only `cgroup_v1.so` gets compiled and `cgroup/v2` fails with a generic "plugin not found".
- **The v2 plugin also needs a running D-Bus system daemon** for its systemd-scope integration (`entrypoint.sh` starts `dbus-daemon --system`) — and since this container has no real systemd, `cgroup.conf` sets `IgnoreSystemd=yes` so Slurm manages cgroups directly instead of asking systemd for a scope.
- **cgroup v2 refuses to delegate a controller from a cgroup with processes directly attached to it** (the "no internal processes" rule). Docker starts every container process directly in the container's own cgroup, so `entrypoint.sh` relocates itself (and everything it spawns) into a child `init/` cgroup first, then enables `memory`/`pids`/`cpu` in `cgroup.subtree_control` at the container's own level — only after that can Slurm's per-job cgroups underneath actually enforce `ConstrainRAMSpace`.
- The container also needs `privileged: true`, `cgroup: host`, and the host's `/sys/fs/cgroup` mounted in (see `docker-compose.yml`) for any of this to be reachable at all.

One nuance discovered along the way: Slurm's own internal OOM classification (`JobState=OUT_OF_MEMORY`, `ExitCode=0:125`) didn't reliably fire in this setup even once real cgroup enforcement was confirmed working (`MaxRSS` dropping to a few MB instead of growing unconstrained) — the kill itself is real, but Slurm's `oom_kill` event-counter polling didn't pick it up here, possibly due to the nested cgroup relocation above. What *does* show up reliably is the OS-level signal death: `ExitCode=137` (a bash-wrapped batch script, the common case) or `ExitSignal=9` (a script that `exec`s the killed process directly) — this is exactly what led to fixing `oomApplies()` to check for those in addition to the textbook `OUT_OF_MEMORY` state.

## Tear down

```
docker compose down
```

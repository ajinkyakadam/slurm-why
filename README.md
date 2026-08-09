# slurm-why

[![CI](https://github.com/ajinkyakadam/slurm-why/actions/workflows/ci.yml/badge.svg)](https://github.com/ajinkyakadam/slurm-why/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-73.3%25-yellow)](https://github.com/ajinkyakadam/slurm-why/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Downloads](https://img.shields.io/github/downloads/ajinkyakadam/slurm-why/total)](https://github.com/ajinkyakadam/slurm-why/releases)
[![DOI](https://zenodo.org/badge/1315577683.svg)](https://doi.org/10.5281/zenodo.21753048)

Plain-English diagnoses for confusing Slurm job states — instead of decoding raw `sacct`/`squeue`/`scontrol` output yourself, run `slurm-why <job-id>` and get a **Diagnosis → Why → Fix**.

If slurm-why saves you a trip to `man squeue`, consider giving it a ⭐ — it helps others on your cluster find it. Run into a confusing Slurm failure it doesn't explain yet? [Open an issue](https://github.com/ajinkyakadam/slurm-why/issues) with the raw `sacct`/`squeue`/`scontrol` output — real day-to-day failure modes are exactly what grows this catalog, and a fix you contribute back helps the next person who hits the same wall.

```
$ slurm-why 483201
Job was killed for using too much memory (OOM)

Why: Your job was killed for exceeding its memory allocation. sacct reports peak
usage (3.9G) below what you requested (4.0G) -- that's not a contradiction: the
cgroup enforcing your memory limit checks continuously, while sacct only samples
roughly every 30 seconds, so a brief spike can trigger the kill without ever
showing up in the accounting record.

Fix: Increase your memory request, e.g. --mem=4.9G.
```

## Read-only, no elevated privileges

slurm-why never writes to Slurm and never needs root, sudo, or `slurmrestd` admin access. Every lookup it does is something you can already run yourself as an ordinary user:

| Command | What it reads | Permission needed |
|---|---|---|
| `sacct` | your job's accounting record (state, exit code, memory usage, time) | none — every user can query their own job history |
| `squeue` | your job's live pending-reason code | none — own jobs are always visible |
| `scontrol show partition` | partition ceilings (max time, nodes, memory) | none — partition config is public, users need it to pick a partition |
| `sacctmgr show qos` / `show assoc` | QOS and association ceilings | none for *viewing* — write access is what's admin-gated |

The one config-dependent caveat: some sites set `PrivateData=jobs` in `slurm.conf`, which restricts visibility of *other users'* jobs. It never blocks you from seeing your own jobs, so it doesn't affect slurm-why's core functionality.

## Install

### RHEL-family: no-root or RPM

Full instructions (checksum verification, air-gapped mirroring) in [`docs/install.md`](docs/install.md). Short version:
```
# no root needed
curl -LO https://github.com/ajinkyakadam/slurm-why/releases/latest/download/slurm-why_linux_amd64.tar.gz
tar xzf slurm-why_linux_amd64.tar.gz && mv slurm-why ~/.local/bin/

# or, cluster-wide
sudo dnf install ./slurm-why-<version>-1.<arch>.rpm
```

### Building from source (works now)

Requires Go 1.23+.

```
git clone https://github.com/ajinkyakadam/slurm-why.git
cd slurm-why
go build -o slurm-why ./cmd/slurm-why
./slurm-why version
```

## Configuration

Today, slurm-why always uses the CLI backend (`sacct`/`squeue`/`scontrol`/`sacctmgr`) — there's nothing to configure, it works out of the box on any login node.

A `slurmrestd` REST backend is planned (see roadmap below). There's no way to auto-discover *where* `slurmrestd` is actually listening — that varies by site (often the login/head node, but sometimes elsewhere to avoid making it a single point of failure), so an admin will need to set slurm-why's own endpoint variable by hand once this backend lands. Falls back to the CLI backend otherwise. Not implemented yet.

## Usage

```
slurm-why <job-id>
```

Add `-v`/`--verbose` to see which data backend was used:
```
slurm-why -v <job-id>
```

### Admin mode

For operators, not just a job's own user:
```
slurm-why --admin <job-id>    # adds a raw technical-detail block (account, partition, QOS, requested resources, exit code) after the diagnosis
slurm-why admin queue         # triages every pending job cluster-wide, bucketing "will never run" vs "just waiting" -- surfaces the broken ones before a confused user files a ticket
```
`--admin` doesn't change what data gets fetched -- `slurm-why <job-id>` already works for any job regardless of owner, since Slurm's own permission model (not this tool) governs who can see what.

## Diagnoses

| Check | Status | Description |
|---|---|---|
| OOM-kill detection | ✅ | Job died `OUT_OF_MEMORY`; explains the sacct-sampling-vs-cgroup nuance; suggests a corrected `--mem` |
| Pending-reason decoder | ✅ | Translates `NODELIST(REASON)` codes into plain English + a fix |
| Impossible-request detector | ✅ | Job requests more than any partition/QOS will ever grant, so it pends forever with no error |
| Dependency failures | ✅ | Names the specific parent job whose outcome broke a `--dependency=afterok` chain |
| Non-zero exit codes decoded | ✅ | Maps common exit codes (126, 127, 139/segfault) to likely causes for `FAILED` jobs |
| Node failures | ✅ | Explains a terminal `NODE_FAIL` as a hardware issue, not the user's fault |
| Efficiency nudges | ✅ | Flags `COMPLETED` jobs that used far less than requested, suggests a right-sized resubmission |
| Stuck-held requeues | 🟡 | A job held after a launch failure (distinct from `NODE_FAIL`) -- couldn't reproduce this against a live cluster yet, deferred rather than guessing |
| Array-job pitfalls, GPU/`--gres` mistakes, fairshare breakdown | 🟡 | Planned, unscheduled |

## Local development

```
go build ./...
go vet ./...
go test ./... -race -cover
```

Tests are entirely fixture-based (see `internal/datasource/cli/testdata/`) — no real Slurm installation is required to run the test suite. For manual validation against a real cluster, see [`test/cluster/`](test/cluster/README.md) — a local Docker-based Slurm 26.05.2 cluster used to confirm each diagnosis (including real cgroup-enforced OOM kills) against actual `sacct`/`squeue`/`scontrol` output, not just fixtures.

## Contributing

Contributions are welcome — bug reports, feature requests, or PRs.
- Open a GitHub issue to report problems or suggest features
- Open a pull request with improvements
- Share how slurm-why has been useful (or not) at your site

## Citation

If you use slurm-why in your work, please consider citing:
```
@software{slurmwhy,
  author = {Kadam, Ajinkya},
  title = {slurm-why: plain-English diagnoses for confusing Slurm job states},
  year = {2026},
  publisher = {GitHub},
  url = {https://github.com/ajinkyakadam/slurm-why},
  doi = {10.5281/zenodo.21753048}
}
```

That DOI always resolves to the latest release. If you need to cite the exact version you used for reproducibility, [slurm-why's Zenodo record](https://doi.org/10.5281/zenodo.21753048) lists a separate, permanent DOI for every tagged release.

## License

MIT

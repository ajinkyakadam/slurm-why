# slurm-why

Plain-English diagnoses for confusing Slurm job states — instead of decoding raw `sacct`/`squeue`/`scontrol` output yourself, run `slurm-why <job-id>` and get a **Diagnosis → Why → Fix**.

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

**Building from source is the only path that works today** — no release has been tagged yet, so the options below aren't live. They're fully built and validated (see `.goreleaser.yaml`, exercised locally with `goreleaser release --snapshot`), just waiting on the first `vX.Y.Z` tag — see `docs/releasing.md` for the versioning plan.

### Homebrew (macOS) — coming soon
```
brew install ajinkyakadam/slurm-why/slurm-why
```

### RHEL-family: no-root or RPM — coming soon
Full instructions (including checksum verification and air-gapped mirroring) in [`docs/install.md`](docs/install.md) once a release exists. Short version:
```
# no root needed
tar xzf slurm-why_<version>_linux_amd64.tar.gz && mv slurm-why ~/.local/bin/

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

A `slurmrestd` REST backend is planned (see roadmap below): when available, it'll be preferred automatically if a site configures an endpoint and JWT token, falling back to the CLI backend otherwise. Not implemented yet.

## Usage

```
slurm-why <job-id>
```

Add `-v`/`--verbose` to see which data backend was used:
```
slurm-why -v <job-id>
```

## Diagnoses

| Check | Status | Description |
|---|---|---|
| OOM-kill detection | ✅ | Job died `OUT_OF_MEMORY`; explains the sacct-sampling-vs-cgroup nuance; suggests a corrected `--mem` |
| Pending-reason decoder | ✅ | Translates `NODELIST(REASON)` codes into plain English + a fix |
| Impossible-request detector | ✅ | Job requests more than any partition/QOS will ever grant, so it pends forever with no error |
| Dependency failures | 🟡 | Point at the failed parent job when `DependencyNeverSatisfied` |
| Non-zero exit codes decoded | 🟡 | Map common exit codes/signals (137, 139, 124, 127) to likely causes |
| Node failures / stuck requeues | 🟡 | `NODE_FAIL`, held-after-requeue jobs |
| Efficiency nudges | 🟡 | Compare requested vs. used resources on completed jobs |
| Array-job pitfalls, GPU/`--gres` mistakes, fairshare breakdown | 🟡 | Planned, unscheduled |

## Local development

```
go build ./...
go vet ./...
go test ./... -race -cover
```

Tests are entirely fixture-based (see `internal/datasource/cli/testdata/`) — no real Slurm installation is required to run the test suite. Manually exercising the CLI against a live cluster is a separate, ongoing effort (tracked internally) since CI can't spin up a real `slurmctld`/`slurmdbd`.

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
  url = {https://github.com/ajinkyakadam/slurm-why}
}
```

## License

MIT

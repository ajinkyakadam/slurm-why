# Releasing slurm-why

## Versioning

Semantic versioning, tags as `vX.Y.Z`. What bumps which number is deliberately tied to what an HPC admin sitting on a login node actually cares about, not just "did the code change":

- **PATCH** (`v0.1.1`): bug fixes only — no new checks, no new reason codes as a feature, no behavior change beyond "this was wrong before and is now correct" (e.g. the `ExitCode=137` OOM-detection fix, or a genuine parsing bug). Always safe to upgrade immediately; nothing an admin needs to review first.
- **MINOR** (`v0.2.0`): new capability, backward compatible — a new Tier 2/3 diagnosis check, a new backend (e.g. the planned `slurmrestd` support), expanded reason-code coverage added as a deliberate feature, a new install method. Existing usage keeps working exactly as before; upgrading only adds behavior, never changes it.
- **MAJOR** (`v1.0.0+`): breaking changes — CLI flag renames/removals, output text/format changes that would break a script parsing `slurm-why`'s output, config format changes. Only meaningful once we actually cross 1.0; while pre-1.0, treat MINOR as the practical ceiling and call out any breaking change explicitly in the release notes even though semver formally permits it in a 0.x minor bump.

No version file to bump by hand — GoReleaser injects the version string from the git tag itself via `-ldflags` (see `.goreleaser.yaml`'s `builds.ldflags`, targeting `internal/cli.version`/`.commit`/`.date`). Tagging is the only step.

## Cutting a release

1. Confirm CI is green on `main`.
2. Decide the bump type per the rules above.
3. `git tag vX.Y.Z && git push origin vX.Y.Z` — this triggers `.github/workflows/release.yml`, which runs GoReleaser to:
   - cross-compile binaries for all 5 targets (linux/{amd64,arm64,ppc64le}, darwin/{amd64,arm64})
   - build RPM packages (NVR-versioned: `slurm-why-X.Y.Z-1.<arch>.rpm`)
   - generate `checksums.txt` for every archive/package
   - create a GitHub Release with an auto-generated changelog from commit messages
4. After GoReleaser finishes, edit the GitHub Release description to add one line: **which live Slurm version(s) this release was validated against** (see `test/cluster/`). This is the single most useful thing an admin sees before deciding to upgrade — everything else in the changelog is "what changed," this is "does it still work with what I actually run."

## What makes this easy for HPC admins specifically

- **Pinned, not floating, installs.** `docs/install.md`'s no-root instructions show both a `/releases/latest/download/...` one-liner (convenient for a quick trial) and a pinned `/releases/download/vX.Y.Z/...` variant. Most sites will want the pinned form for the same reason this repo pins its own Docker base images and GitHub Actions: a login-node install shouldn't change out from under anyone just because upstream cut a new tag.
- **Checksum verification is one `sha256sum -c` away.** GoReleaser publishes `checksums.txt` alongside every release; `docs/install.md` shows verifying against it before running anything downloaded onto a shared login node.
- **RPM versioning is just normal `dnf` semantics.** Name-version-release packaging means an admin can mirror multiple versions in an internal repo and use `dnf upgrade`/`dnf downgrade` exactly like any other package — no slurm-why-specific tooling to learn.
- **Air-gapped mirroring is straightforward.** A pinned download + its checksum are exactly what's needed to copy a known-good artifact onto an isolated cluster's internal mirror once, with no repeated trust decisions per node.

## Pre-release checklist

- [ ] CI green on `main`
- [ ] `test/cluster/` still validates the 3 v0.1 diagnoses (rebuild + rerun if anything in `internal/datasource` or `internal/diagnose` changed since the last release)
- [ ] Bump type matches the rules above, not just "whatever feels right"
- [ ] After tagging: edit the GitHub Release description to note the Slurm version(s) validated against

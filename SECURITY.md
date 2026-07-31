# Security Policy

## Supported versions

Only the latest released version is supported. Since slurm-why has no persistent state or server component, upgrading is always a matter of dropping in a new binary (see [`docs/install.md`](docs/install.md)) -- there's no reason to maintain fixes on older versions.

## Reporting a vulnerability

Please report security issues privately via [GitHub's private vulnerability reporting](https://github.com/ajinkyakadam/slurm-why/security/advisories/new) rather than a public issue, so there's a window to fix and release before details are public.

If you're not sure whether something qualifies, err on the side of reporting it privately -- worst case, it gets reclassified as a normal issue.

## What's in scope

slurm-why is a read-only CLI: it shells out to `sacct`/`squeue`/`scontrol`/`sacctmgr` (via `exec.Command` with an argument list, never a shell string, so there's no command-injection surface from job IDs or other input) and parses their output. It never writes to Slurm and needs no elevated privileges. The realistic security surface is narrow, but still worth reporting if found:

- A crafted `sacct`/`squeue`/`scontrol` output value (e.g. a malicious job name or comment field) that causes a crash, panic, or incorrect behavior when parsed.
- Any code path that shells out with unsanitized input in a way that *could* become injectable.
- Once the planned `slurmrestd` REST backend lands: any mishandling of the JWT auth token (e.g. logging it, sending it somewhere unintended).

## What's out of scope

- Vulnerabilities in Slurm itself -- report those to [SchedMD](https://www.schedmd.com/) directly.
- The fact that slurm-why can see any job data a user's own Slurm permissions already allow (see the README's "Read-only, no elevated privileges" section) -- that's Slurm's access-control model, not something slurm-why changes.

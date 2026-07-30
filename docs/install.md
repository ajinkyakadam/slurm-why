# Installing slurm-why on RHEL-family systems

Two paths, depending on who's installing and where.

## Unprivileged user, no root (works on any login node)

Most login-node users can't `sudo`. This path needs no package manager and no network access on the login node itself — download elsewhere and copy the binary over if the node is air-gapped.

**Pick a version.** Either the latest release:

```
curl -LO https://github.com/ajinkyakadam/slurm-why/releases/latest/download/slurm-why_linux_amd64.tar.gz
curl -LO https://github.com/ajinkyakadam/slurm-why/releases/latest/download/checksums.txt
```

or a specific, pinned version — recommended if you want your login-node install to only change when you decide to bump it, not silently whenever a new release ships:

```
VERSION=0.1.0
curl -LO https://github.com/ajinkyakadam/slurm-why/releases/download/v${VERSION}/slurm-why_${VERSION}_linux_amd64.tar.gz
curl -LO https://github.com/ajinkyakadam/slurm-why/releases/download/v${VERSION}/checksums.txt
```

(swap `amd64` for `arm64` on ARM login nodes)

**Verify the checksum** before running anything downloaded onto a shared system:

```
sha256sum --ignore-missing -c checksums.txt
```

**Install:**

```
tar xzf slurm-why_*_linux_amd64.tar.gz
mkdir -p ~/.local/bin
mv slurm-why ~/.local/bin/
slurm-why version
```

**Air-gapped login nodes**: run the download+verify steps on any machine with internet access, then `scp` the tarball and `checksums.txt` over and verify again on the target before installing — no package manager or outbound network access needed on the login node itself.

## Sysadmin, cluster-wide (RPM)

```
sudo dnf install ./slurm-why-<version>-1.<arch>.rpm
```

RPMs are name-version-release packaged (`slurm-why-0.1.0-1.x86_64.rpm`), so normal `dnf` semantics apply — `dnf upgrade`/`dnf downgrade` work exactly like any other package. Multiple versions can be mirrored in an internal repo for controlled rollout across a cluster, which is also the natural way to handle air-gapped sites: mirror the RPM (and its checksum from `checksums.txt`) into your site's internal yum repo once, and every node installs from there without ever reaching GitHub directly.

## Verify

```
slurm-why version
slurm-why --help
```

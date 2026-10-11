#!/usr/bin/env bash
# Build slurm-why-<version>.tar.gz: the tagged source tree plus a vendor/
# directory holding every Go module dependency, so distro packagers
# (OpenHPC's OBS, Fedora/EPEL's Koji) can `go build -mod=vendor` with no
# network access. vendor/ is generated here at release time and never
# committed to the repo.
#
# Usage: scripts/vendored-source-tarball.sh <version> <output-dir>
#
# Run from the repo root by GoReleaser's before hook (see .goreleaser.yaml).
# Requires GNU tar for the reproducibility flags.
set -euo pipefail

version="${1:?usage: $0 <version> <output-dir>}"
outdir="${2:?usage: $0 <version> <output-dir>}"
name="slurm-why-${version}"

work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT

# Start from exactly what's committed at HEAD, not the working tree, so
# stray local files can never leak into a release tarball.
git archive --format=tar --prefix="${name}/" HEAD | tar -x -C "${work}"
(cd "${work}/${name}" && go mod vendor)

# Pin timestamps/ownership/order to the commit so rebuilding the same
# tag yields a byte-identical tarball (and a stable checksum).
mtime="$(git log -1 --format=%ct HEAD)"
mkdir -p "${outdir}"
tar -C "${work}" \
	--sort=name --mtime="@${mtime}" \
	--owner=0 --group=0 --numeric-owner \
	-cf - "${name}" | gzip -n >"${outdir}/${name}.tar.gz"

echo "wrote ${outdir}/${name}.tar.gz"

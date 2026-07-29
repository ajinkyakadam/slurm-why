# Pinned to a specific official Ubuntu 22.04 image digest (not a floating
# tag) so the base is reproducible and its provenance is verifiable.
FROM ubuntu@sha256:0e0a0fc6d18feda9db1590da249ac93e8d5abfea8f4c3c0c849ce512b5ef8982

ENV DEBIAN_FRONTEND=noninteractive

# Build/runtime dependencies, all from Ubuntu's own signed apt repos.
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential ca-certificates wget \
    munge libmunge-dev \
    mariadb-server libmariadb-dev \
    libssl-dev libpam0g-dev python3 perl pkg-config \
    libdbus-1-dev libhwloc-dev dbus \
    procps gosu \
    && rm -rf /var/lib/apt/lists/*

# Slurm 26.05.2 -- built from SchedMD's own official GitHub release
# (github.com/SchedMD/slurm), verified against the SHA256 digest GitHub's
# release API reports for this exact asset, not a distro package (no
# distro ships anything close to current) and not a third-party image.
ARG SLURM_VERSION=26.05.2
ARG SLURM_RELEASE_TAG=slurm-26-05-2-1
ARG SLURM_SHA256=01151f233dde64d95cc4ca96931aea3d4a2cad372ed9ad86a8b9f6b674600c3b
RUN wget -q -O /tmp/slurm.tar.bz2 \
      "https://github.com/SchedMD/slurm/releases/download/${SLURM_RELEASE_TAG}/slurm-${SLURM_VERSION}.tar.bz2" \
    && echo "${SLURM_SHA256}  /tmp/slurm.tar.bz2" | sha256sum -c - \
    && mkdir -p /usr/src/slurm \
    && tar -xjf /tmp/slurm.tar.bz2 -C /usr/src/slurm --strip-components=1 \
    && rm /tmp/slurm.tar.bz2

RUN cd /usr/src/slurm \
    && ./configure --prefix=/usr --sysconfdir=/etc/slurm \
    && make -j"$(nproc)" \
    && make install \
    && ldconfig

RUN useradd -r -s /usr/sbin/nologin slurm \
    && useradd -m -s /bin/bash testuser \
    && mkdir -p /etc/slurm /var/spool/slurmctld /var/spool/slurmd /var/log/slurm /var/run/slurm \
    && chown slurm:slurm /var/spool/slurmctld /var/log/slurm /var/run/slurm

COPY entrypoint.sh /entrypoint.sh
COPY slurm.conf /etc/slurm/slurm.conf
COPY slurmdbd.conf /etc/slurm/slurmdbd.conf
COPY cgroup.conf /etc/slurm/cgroup.conf
RUN chmod +x /entrypoint.sh \
    && chmod 600 /etc/slurm/slurmdbd.conf \
    && chown slurm:slurm /etc/slurm/slurmdbd.conf

ENTRYPOINT ["/entrypoint.sh"]

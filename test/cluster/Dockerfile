# Pinned to a specific official Ubuntu 22.04 image digest (not a floating
# tag) so the base is reproducible and its provenance is verifiable.
FROM ubuntu@sha256:0e0a0fc6d18feda9db1590da249ac93e8d5abfea8f4c3c0c849ce512b5ef8982

ENV DEBIAN_FRONTEND=noninteractive

# All packages below come from Ubuntu's own signed apt repositories --
# no third-party Docker images, no compiling Slurm from source.
RUN apt-get update && apt-get install -y --no-install-recommends \
    slurmctld slurmd slurmdbd slurm-client \
    munge \
    mariadb-server \
    procps \
    gosu \
    && rm -rf /var/lib/apt/lists/*

# testuser must exist as a real OS account -- Slurm runs job steps as the
# submitting user's actual UID, matching the sacctmgr association created
# in entrypoint.sh.
RUN useradd -m -s /bin/bash testuser

COPY entrypoint.sh /entrypoint.sh
COPY slurm.conf /etc/slurm/slurm.conf
COPY slurmdbd.conf /etc/slurm/slurmdbd.conf
RUN chmod +x /entrypoint.sh \
    && chmod 600 /etc/slurm/slurmdbd.conf \
    && chown slurm:slurm /etc/slurm/slurmdbd.conf

ENTRYPOINT ["/entrypoint.sh"]

#!/bin/bash
set -e

echo "=== starting munge ==="
service munge start

echo "=== initializing/starting mariadb ==="
if [ ! -d /var/lib/mysql/mysql ]; then
  mariadb-install-db --user=mysql --datadir=/var/lib/mysql >/tmp/mysql-install.log 2>&1
fi
service mariadb start

for i in $(seq 1 30); do
  mysqladmin ping >/dev/null 2>&1 && break
  sleep 1
done

echo "=== configuring accounting database ==="
mysql -e "CREATE DATABASE IF NOT EXISTS slurm_acct_db;"
mysql -e "CREATE USER IF NOT EXISTS 'slurm'@'localhost' IDENTIFIED BY 'slurmdbpass';"
mysql -e "GRANT ALL PRIVILEGES ON slurm_acct_db.* TO 'slurm'@'localhost';"
mysql -e "FLUSH PRIVILEGES;"

mkdir -p /var/spool/slurmctld /var/spool/slurmd /var/log/slurm /var/run/slurm
chown slurm:slurm /var/spool/slurmctld /var/log/slurm /var/run/slurm

echo "=== starting slurmdbd ==="
gosu slurm /usr/sbin/slurmdbd
sleep 3
cat /var/log/slurm/slurmdbd.log || true

echo "=== registering test cluster/account/user ==="
sacctmgr -i add cluster testcluster || true
sacctmgr -i add account testacct Description="test account" Organization="test" || true
sacctmgr -i add user testuser Account=testacct || true

echo "=== starting slurmctld ==="
/usr/sbin/slurmctld
sleep 2
cat /var/log/slurm/slurmctld.log || true

echo "=== starting slurmd ==="
/usr/sbin/slurmd
sleep 2
cat /var/log/slurm/slurmd.log || true

echo "=== ready ==="
tail -f /dev/null

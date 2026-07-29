#!/bin/bash
#SBATCH --job-name=normal
#SBATCH --time=00:01:00
#SBATCH --mem=100M
#SBATCH --output=/home/testuser/normal-%j.out
echo "hello from a normal job"
sleep 2

#!/bin/bash
#SBATCH --job-name=impossible_time
#SBATCH --time=02:00:00
#SBATCH --mem=100M
#SBATCH --output=/home/testuser/impossible_time-%j.out
echo "this should never run -- the partition's MaxTime=01:00:00"

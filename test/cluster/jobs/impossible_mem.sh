#!/bin/bash
#SBATCH --job-name=impossible_mem
#SBATCH --time=00:01:00
#SBATCH --mem=3000M
#SBATCH --output=/home/testuser/impossible_mem-%j.out
echo "this should never run -- the partition's MaxMemPerNode=2000"

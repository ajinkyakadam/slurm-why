#!/bin/bash
#SBATCH --job-name=impossible_nodes
#SBATCH --time=00:01:00
#SBATCH --mem=100M
#SBATCH --nodes=2
#SBATCH --output=/home/testuser/impossible_nodes-%j.out
echo "this should never run -- the partition's MaxNodes=1"

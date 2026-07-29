#!/bin/bash
#SBATCH --job-name=oom
#SBATCH --time=00:02:00
#SBATCH --mem=100M
#SBATCH --output=/home/testuser/oom-%j.out
python3 -c "
data = []
try:
    while True:
        data.append(bytearray(10*1024*1024))  # grow 10MB at a time until killed
except MemoryError:
    pass
"

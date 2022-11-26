#!/bin/bash

N_EXPERIMENTS=10
N_DISKS=$(sudo fdisk -l /dev/nvme[0-9]n[0-9] | grep /dev | wc -l )
N_CONCURRENT=$(($N_DISKS - 1))
#free slots = N_CONCURRENT

start source
sync source

for experiments:
    get lock
    if got lock
        copy state
        docker run (target, nvme path, error model)
        release lock
    else
        wait
    

echo $N_CONCURRENT
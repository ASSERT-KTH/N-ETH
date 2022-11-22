#!/bin/bash

N_DISKS=$(sudo fdisk -l /dev/nvme[0-9]n[0-9] | grep /dev | wc -l )

N_CONCURRENT=(($N_DISKS - 1))

echo $N_CONCURRENT
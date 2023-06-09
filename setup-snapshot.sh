#!/bin/bash
set -x

# ssd setup
# find device name
echo "Finding snapshot partition name..."
SSD_PARTITION=`lsblk | grep 1.5T | awk -e '$0 ~ /sd.\s/ {print $1}'`
SSD_PARTITION=/dev/"$SSD_PARTITION"1
echo "Partition name found : $SSD_PARTITION"

# umount
SSD_MOUNT_POINT=/home/javier/ssd
sudo umount $SSD_MOUNT_POINT

NVME_PARTITION=/dev/nvme0n1p1

# copy eth state from snapshot ? maybe copy directory instead of partition
sudo dd if=$SSD_PARTITION of=$NVME_PARTITION bs=500M status=progress

# mount nvme
NVME_MOUNT_POINT=/home/javier/nvme
sudo mount $NVME_PARTITION $NVME_MOUNT_POINT

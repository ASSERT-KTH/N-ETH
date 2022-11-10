#!/bin/bash
set -x

# ssd setup
# find device name
echo "Finding snapshot partition name..."
SSD_PARTITION=`lsblk | grep 1.5T | awk -e '$0 ~ /sd.\s/ {print $1}'`
SSD_PARTITION=/dev/"$SSD_PARTITION"1
echo "Partition name found : $SSD_PARTITION"

# mount
SSD_MOUNT_POINT=/home/javier/ssd
sudo mount $SSD_PARTITION $SSD_MOUNT_POINT

# nvme setup
# nvme create partition
NVME_DEVICE=/dev/nvme0n1
sudo fdisk $NVME_DEVICE <<EOF
n
p
1


w
EOF

# format new partition
NVME_PARTITION=/dev/nvme0n1p1
sudo mkfs.ext4 -F $NVME_PARTITION

# copy eth state from snapshot ? maybe copy directory instead of partition
sudo dd if=$SSD_PARTITION of=$NVME_PARTITION bs=500M status=progress

# mount nvme
NVME_MOUNT_POINT=/home/javier/nvme
sudo mount $NVME_PARTITION $NVME_MOUNT_POINT

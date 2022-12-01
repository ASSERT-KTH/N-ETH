#!/bin/bash
set -x

# nvme setup
NVMES=($(sudo fdisk -l /dev/nvme[0-9]n[0-9] | grep /dev | awk '{print substr($2, 1, length($2)-1) }'))

for NVME_DEVICE in "${NVMES[@]}"; do
   sudo fdisk $NVME_DEVICE <<EOF
n
p
1


w
EOF

sudo mkfs.ext4 -F "$NVME_DEVICE"p1
done

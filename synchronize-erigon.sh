#!/bin/bash

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

# start erigon
ERIGON_DIR=/home/javier/erigon/build/bin
ERIGON_LOG=/home/javier/erigon-sync-"$(date -I)".log
cd $ERIGON_DIR
{ ./erigon --datadir=/home/javier/nvme/data-dir --http &> $ERIGON_LOG; } &
sleep 2
ERIGON_PID=`ps aux | grep "\\./[e]rigon" | awk '{print $2}'`

# start teku
TEKU_DIR=/home/javier/teku/build/install/teku/bin
TEKU_LOG=/home/javier/teku-sync-"$(date -I)".log
cd $TEKU_DIR
{ ./teku --ee-endpoint=http://localhost:8551 --ee-jwt-secret-file=/home/javier/nvme/data-dir/jwt.hex --data-beacon-path=/home/javier/nvme/teku-data-dir/ &> $TEKU_LOG; } &
sleep 2
TEKU_PID=`ps aux | grep "teku\\.home" | awk '{print $2}'`

# check is synchonized < 2 blocks from etherscan
SYNC_DISTANCE=10000

while [ $SYNC_DISTANCE -gt 2 ]
do
    # wait 30 seconds
    sleep 30

    # curl to etherscan
    ETHERSCAN_BLOCK_HEX=`curl 'https://api.etherscan.io/api?module=proxy&action=eth_blockNumber' | jq -r .result | awk '{ print substr( $0, 3 ) }' | awk '{print toupper($0)}'`
    ETHERSCAN_BLOCK=`echo "obase=10; ibase=16; $ETHERSCAN_BLOCK_HEX" | bc`

    # curl to erigon and get number
    ERIGON_BLOCK_HEX=`curl --data '{"method":"eth_getBlockByNumber","params":["latest", false],"id":1,"jsonrpc":"2.0"}' -H "Content-Type: application/json" -X POST 127.0.0.1:8545 | jq -r .result.number | awk '{ print substr( $0, 3 ) }' | awk '{print toupper($0)}'`
    ERIGON_BLOCK=`echo "obase=10; ibase=16; $ERIGON_BLOCK_HEX" | bc`

    # compute distances
    SYNC_DISTANCE=$(( $ETHERSCAN_BLOCK - $ERIGON_BLOCK ))
    echo $SYNC_DISTANCE
done


# save logs
# set connection string
# AZURE_STORAGE_CONNECTION_STRING==xxxxxx

# erigon
# az storage blob upload -f $ERIGON_LOG -c logs -n $ERIGON_LOG --connection-string="$AZURE_STORAGE_CONNECTION_STRING"

# teku
# az storage blob upload -f $ERIGON_LOG -c logs -n $ERIGON_LOG --connection-string="$AZURE_STORAGE_CONNECTION_STRING"

# stop erigon

kill -2 $ERIGON_PID
# stop teku
kill -2 $TEKU_PID

# check that processes have terminated.
ERIGON_GREP="erigon"
TEKU_GREP="teku"

while [ ! -z $ERIGON_GREP ] || [ ! -z $TEKU_GREP ]
do
    sleep 10
    ERIGON_GREP=`ps aux | grep "\\./[g]eth"`
    TEKU_GREP=`ps aux | grep "teku\\.home"`
done

echo "Sync complete!"

# copy eth state to snapshot

# rsync --delete -r nvme/ ssd

# # save snapshot

# az snapshot this disk 

# # create new version of vm

# az image this vm

# # delete disk

# unmount ssd disk
# az delete disk

# # shut down
# shutdown now

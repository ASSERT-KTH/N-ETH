#!/bin/bash

# ssd setup

# find device name
DEVICE=`lsblk | grep 1.5T | awk -e '$0 ~ /sd.\s/ {print $1}'`
DEVICE="$DEVICE"1
# mount
sudo mount /dev/$DEVICE /home/javier/ssd

# nvme setup
# nvme create partition

sudo fdisk -u -p /dev/nvme0n1 <<EOF
n
p
1

w
EOF

# copy eth state from snapshot ? maybe copy directory instead of partition
sudo dd if=/dev/sda1 of=/dev/nvme0n1p1 bs=500M status=progress

# mount nvme
sudo mount /dev/nvme0n1p1 /home/javier/nvme

# start geth
cd /home/javier/go-ethereum/build/bin
GETH_LOG=`/home/javier/geth-sync-"$(date -I)".log`
nohup ./geth --datadir=/home/javier/nvme/data-dir -http > $GETH_LOG &
GETH_PID=$!

# start teku
cd /home/javier/teku/build/install/teku/bin
TEKU_LOG=`/home/javier/teku-sync"$(date -I)".log`
nohup ./teku --ee-endpoint=http://localhost:8551 --ee-jwt-secret-file=/home/javier/nvme/data-dir/geth/jwtsecret --data-beacon-path=/home/javier/nvme/teku-data-dir/ > $TEKU_LOG &
TEKU_PID=$!

# check is synchonized < 2 blocks from etherscan
SYNC_DISTANCE=10000

while [ $SYNC_DISTANCE -gt 2 ]
do
    # wait 30 seconds
    sleep 30

    # curl to etherscan
    ETHERSCAN_BLOCK_HEX=`curl 'https://api.etherscan.io/api?module=proxy&action=eth_blockNumber' | jq -r .result | awk '{ print substr( $0, 3 ) }' | awk '{print toupper($0)}'`
    ETHERSCAN_BLOCK=`echo "obase=10; ibase=16; $ETHERSCAN_BLOCK_HEX" | bc`

    # curl to geth and get number
    GETH_BLOCK_HEX=`curl --data '{"method":"eth_getBlockByNumber","params":["latest", false],"id":1,"jsonrpc":"2.0"}' -H "Content-Type: application/json" -X POST 127.0.0.1:8545 | jq -r .result.number | awk '{ print substr( $0, 3 ) }' | awk '{print toupper($0)}'`
    GETH_BLOCK=`echo "obase=10; ibase=16; $ETHERSCAN_BLOCK_HEX" | bc`

    # compute distances
    SYNC_DISTANCE=$(( $ETHERSCAN_BLOCK - $GETH_BLOCK ))
    echo $SYNC_DISTANCE
done


# save logs
# set connection string
AZURE_STORAGE_CONNECTION_STRING==xxxxxx

# geth
az storage blob upload -f geth.log -c logs -n $GETH_LOG

# teku
az storage blob upload -f geth.log -c logs -n $TEKU_LOG

# stop geth

kill -2 GETH_PID

# stop teku

kill -2 TEKU_PID

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

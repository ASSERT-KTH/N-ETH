#!/bin/bash

# start geth
GETH_DIR=/home/javier/go-ethereum/build/bin
GETH_LOG=/home/javier/geth-sync-"$(date -I)".log
cd $GETH_DIR
{ ./geth --datadir=/home/javier/nvme/data-dir -http &> $GETH_LOG; } &
sleep 2
GETH_PID=`ps aux | grep "\\./[g]eth" | awk '{print $2}'`

# start teku
TEKU_DIR=/home/javier/teku/build/install/teku/bin
TEKU_LOG=/home/javier/teku-sync-"$(date -I)".log
cd $TEKU_DIR
{ ./teku --ee-endpoint=http://localhost:8551 --ee-jwt-secret-file=/home/javier/nvme/data-dir/geth/jwtsecret --data-beacon-path=/home/javier/nvme/teku-data-dir/ &> $TEKU_LOG; } &
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

    # curl to geth and get number
    GETH_BLOCK_HEX=`curl --data '{"method":"eth_getBlockByNumber","params":["latest", false],"id":1,"jsonrpc":"2.0"}' -H "Content-Type: application/json" -X POST 127.0.0.1:8545 | jq -r .result.number | awk '{ print substr( $0, 3 ) }' | awk '{print toupper($0)}'`
    GETH_BLOCK=`echo "obase=10; ibase=16; $GETH_BLOCK_HEX" | bc`

    # compute distances
    SYNC_DISTANCE=$(( $ETHERSCAN_BLOCK - $GETH_BLOCK ))
    echo $SYNC_DISTANCE
done


# save logs
# set connection string
# AZURE_STORAGE_CONNECTION_STRING==xxxxxx

# geth
# az storage blob upload -f $GETH_LOG -c logs -n $GETH_LOG --connection-string="$AZURE_STORAGE_CONNECTION_STRING"

# teku
# az storage blob upload -f $GETH_LOG -c logs -n $TEKU_LOG --connection-string="$AZURE_STORAGE_CONNECTION_STRING"

# stop geth
kill -2 $GETH_PID

# stop teku
kill -2 $TEKU_PID

# check that processes have terminated.
GETH_GREP="geth"
TEKU_GREP="teku"

while [ ! -z $GETH_GREP ] || [ ! -z $TEKU_GREP ]
do
    sleep 10
    GETH_GREP=`ps aux | grep "\\./[g]eth"`
    TEKU_GREP=`ps aux | grep "teku\\.home"`
done

echo "Sync complete!"

# copy eth state to snapshot

rsync --delete -r nvme/ ssd

# # save snapshot

# az snapshot this disk 

# # create new version of vm

# az image this vm

# # delete disk

# unmount ssd disk
# az delete disk

# # shut down
# shutdown now

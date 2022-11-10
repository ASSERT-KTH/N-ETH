#!/bin/bash
set -x

# start besu
BESU_DIR=/home/javier/besu/build/install/besu/bin
BESU_LOG=/home/javier/besu-sync-"$(date -I)".log
cd $BESU_DIR
{ ./besu --rpc-http-enabled --data-path=/home/javier/nvme/data-dir --pruning-enabled &> $BESU_LOG; } &
sleep 2
BESU_PID=`ps aux | grep "besu\\.home" | awk '{print $2}'`

# wait for jwt key
STAT_RETURN=1
while [ $STAT_RETURN -ne 0 ]
do
    stat /home/javier/nvme/data-dir/jwt.hex
    STAT_RETURN=$?
    sleep 10
done

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

    # curl to besu and get number
    BESU_BLOCK_HEX=`curl --data '{"method":"eth_getBlockByNumber","params":["latest", false],"id":1,"jsonrpc":"2.0"}' -H "Content-Type: application/json" -X POST 127.0.0.1:8545 | jq -r .result.number | awk '{ print substr( $0, 3 ) }' | awk '{print toupper($0)}'`
    BESU_BLOCK=`echo "obase=10; ibase=16; $BESU_BLOCK_HEX" | bc`

    # compute distances
    SYNC_DISTANCE=$(( $ETHERSCAN_BLOCK - $BESU_BLOCK ))
    echo $SYNC_DISTANCE
done


# save logs
# set connection string
# AZURE_STORAGE_CONNECTION_STRING==xxxxxx

# besu
# az storage blob upload -f $BESU_LOG -c logs -n $BESU_LOG

# teku
# az storage blob upload -f $TEKU_LOG -c logs -n $TEKU_LOG

# stop besu
kill -2 $BESU_PID

# stop teku
kill -2 $TEKU_PID

# check that processes have terminated.
BESU_GREP="besu"
TEKU_GREP="teku"

while [ ! -z "$BESU_GREP" ] || [ ! -z "$TEKU_GREP" ]
do
    sleep 10
    BESU_GREP=`ps aux | grep "besu\\.home"`
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
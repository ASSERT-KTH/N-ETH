#!/bin/bash
set -x

CONFIG_FILE=$(pwd)/config.toml
TARGET=$1

get_config () {
    stoml $CONFIG_FILE $1
}

# if target is valid client

# get working dir
WORKING_DIR=$(get_config "working_dir")

# start target
TARGET_DIR=$(get_config "$TARGET.exec_dir")
TARGET_LOG="$WORKING_DIR/$TARGET-sync-\"$(date -I)\".log"
cd $TARGET_DIR
TARGET_CMD=$(get_config "$TARGET.exec_cmd")
{ $TARGET_CMD &> $TARGET_LOG; } &
sleep 2
TARGET_GREP_STR=$(get_config "$TARGET.grep_str")
TARGET_PID=`ps aux | grep "$TARGET_GREP_STR" | awk '{print $2}'`

# start teku
TEKU_DIR=/home/javier/teku/build/install/teku/bin
TEKU_LOG=/home/javier/teku-sync-"$(date -I)".log
cd $TEKU_DIR
TARGET_JWT_FILE=$(get_config "$TARGET.jwt_path")
{ ./teku --ee-endpoint=http://localhost:8551 --ee-jwt-secret-file=$TARGET_JWT_FILE --data-beacon-path=/home/javier/nvme/teku-data-dir/ &> $TEKU_LOG; } &
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

    # curl to target and get number
    TARGET_BLOCK_HEX=`curl --data '{"method":"eth_getBlockByNumber","params":["latest", false],"id":1,"jsonrpc":"2.0"}' -H "Content-Type: application/json" -X POST 127.0.0.1:8545 | jq -r .result.number | awk '{ print substr( $0, 3 ) }' | awk '{print toupper($0)}'`
    TARGET_BLOCK=`echo "obase=10; ibase=16; $TARGET_BLOCK_HEX" | bc`

    # compute distances
    SYNC_DISTANCE=$(( $ETHERSCAN_BLOCK - $TARGET_BLOCK ))
    echo $SYNC_DISTANCE
done


# save logs
# set connection string
# AZURE_STORAGE_CONNECTION_STRING==xxxxxx

# target
# az storage blob upload -f $TARGET_LOG -c logs -n $TARGET_LOG --connection-string="$AZURE_STORAGE_CONNECTION_STRING"

# teku
# az storage blob upload -f $TARGET_LOG -c logs -n $TARGET_LOG --connection-string="$AZURE_STORAGE_CONNECTION_STRING"

# stop target
kill -2 $TARGET_PID

# stop teku
kill -2 $TEKU_PID

# check that processes have terminated.
TARGET_GREP="target"
TEKU_GREP="teku"

while [ ! -z "$TARGET_GREP" ] || [ ! -z "$TEKU_GREP" ]
do
    sleep 10
    TARGET_GREP=`ps aux | grep "$TARGET_GREP_STR"`
    TEKU_GREP=`ps aux | grep "teku\\.home"`
done

echo "Sync complete!"



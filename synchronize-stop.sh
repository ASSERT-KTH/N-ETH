#!/bin/bash
set -x

CONFIG_FILE=$(pwd)/config.toml
TARGET=$1

get_config () {
    stoml $CONFIG_FILE $1
}

# get working dir
WORKING_DIR=$HOME
OUTPUT_DIR=$(get_config "output_dir")

# start target
TARGET_LOG="$OUTPUT_DIR/$TARGET-sync-$(date -Iseconds).log"
TARGET_CMD=$(get_config "$TARGET.exec_cmd")
DATA_DIR_PARAM=$(get_config "$TARGET.datadir_flag")=$WORKING_DIR/$(get_config "$TARGET.datadir")

JWT_FLAG=$(get_config "$TARGET.jwt_flag")
TARGET_JWT_FILE=$WORKING_DIR/$(get_config "$TARGET.jwt_path")
if [ ! -z JWT_FLAG ]; then
    JWT_PARAM="$JWT_FLAG=$TARGET_JWT_FILE"
fi 

{ $TARGET_CMD $JWT_PARAM $DATA_DIR_PARAM &> $TARGET_LOG; } &
TARGET_PPID=$!
sleep 5
TARGET_GREP_STR=$TARGET_PPID.*$(get_config "$TARGET.grep_str")
TARGET_PID=`ps axo pid,ppid,cmd | grep "$TARGET_GREP_STR" | awk '{print $1}'`

# start teku
TEKU_LOG=$OUTPUT_DIR/teku-sync-$(date -Iseconds).log
{ teku --ee-endpoint=http://localhost:8551 --ee-jwt-secret-file=$TARGET_JWT_FILE --data-beacon-path=$WORKING_DIR/nvme/teku-data-dir/ &> $TEKU_LOG; } &
TEKU_PPID=$!
sleep 2
TEKU_GREP_STR=$TEKU_PPID.*teku\\.home
TEKU_PID=`ps axo pid,ppid,cmd | grep $TEKU_GREP_STR | awk '{print $1}'`

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
    echo "Sync distance: $SYNC_DISTANCE"
done

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
    TARGET_GREP=`ps axo pid,ppid,cmd | grep "$TARGET_GREP_STR"`
    TEKU_GREP=`ps axo pid,ppid,cmd | grep "$TEKU_GREP_STR"`
done

echo "Sync complete!"



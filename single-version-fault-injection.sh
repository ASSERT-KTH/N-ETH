#!/bin/bash
set -x

CONFIG_FILE=$(pwd)/config.toml
TARGET=$1

get_config () {
    stoml $CONFIG_FILE $1
}

WORKING_DIR=$HOME
CHAOS_ETH_DIR=$WORKING_DIR/$(get_config "chaos_eth_dir")
ERROR_MODELS=$CHAOS_ETH_DIR/experiments/$TARGET/error-model.json

# spawn + sync wait
./synchronize $TARGET
echo "START" > ipc.dat

while true
do
    # get working dir

    # start target
    TARGET_LOG="$WORKING_DIR/$TARGET-sync-$(date -I).log"
    TARGET_CMD=$(get_config "$TARGET.exec_cmd")
    DATA_DIR_PARAM=$(get_config "$TARGET.datadir_flag")=$WORKING_DIR/$(get_config "$TARGET.datadir")
    { $TARGET_CMD $DATA_DIR_PARAM &> $TARGET_LOG; } &
    sleep 2
    TARGET_GREP_STR=$(get_config "$TARGET.grep_str")
    TARGET_PID=`ps aux | grep "$TARGET_GREP_STR" | awk '{print $2}'`

    # start teku
    TEKU_LOG=$WORKING_DIR/teku-sync-$(date -I).log
    TARGET_JWT_FILE=$WORKING_DIR/$(get_config "$TARGET.jwt_path")
    { teku --ee-endpoint=http://localhost:8551 --ee-jwt-secret-file=$TARGET_JWT_FILE --data-beacon-path=$WORKING_DIR/nvme/teku-data-dir/ &> $TEKU_LOG; } &
    sleep 2
    TEKU_PID=`ps aux | grep "teku\\.home" | awk '{print $2}'`

    #attach error injection
    sleep 10
    CHAOS_ETH_GREP_STR="[s]yscall_injector.py"
    cd $CHAOS_ETH_DIR
    { python syscall_injector.py --config $ERROR_MODELS -p $TARGET_PID } &
    CHAOS_ETH_PID=`ps aux | grep "$CHAOS_ETH_GREP_STR" | awk '{print $2}'`

    sleep 3

    # check that everything is still running

    TARGET_GREP="target"
    TEKU_GREP="teku"
    CHAOS_ETH_GREP="chaos-eth"

    while [ -z "$TARGET_GREP" ] && [ -z "$TEKU_GREP" ] && [ -z "$CHAOS_ETH_GREP" ]
    do
        TARGET_GREP=`ps aux | grep "$TARGET_GREP_STR"`
        TEKU_GREP=`ps aux | grep "teku\\.home"`
        CHAOS_ETH_GREP=`ps aux | grep "$CHAOS_ETH_GREP"`

        sleep 10
    done

    # if one crashed restart all

    kill -2 $TARGET_PID
    kill -2 $TEKU_PID
    kill -2 $CHAOS_ETH_PID

done
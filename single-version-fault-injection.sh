#!/bin/bash
set -x

get_config () {
    stoml $CONFIG_FILE $1
}

CHAOS_ETH_DIR=$(get_config "chaos_eth_dir")
ERROR_MODELS=./error-models.json
CONFIG_FILE=./config.toml
TARGET=$1


# spawn + sync wait
./synchronize $TARGET
echo "START" > ipc.dat

while true
do
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

    #attach error injection
    CHAOS_ETH_GREP_STR="[s]yscall_injector.py"
    sleep 10
    cd $CHAOS_ETH_DIR
    { python syscall_injector.py -c $ERROR_MODELS -p $TARGET_PID } &
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
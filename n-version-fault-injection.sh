#!/bin/bash
set -x

trap forced_shutdown SIGINT SIGQUIT SIGTERM

forced_shutdown()
{
    exit
}


# TODO:
# exclude sudo if in container
USER=$(whoami)
SUDO=""
if [[ $USER != "root" ]]; then
    SUDO="sudo"
else
    # HACK: root if inside container -> mount debugfs and tracefs
    mount -t debugfs debugfs /sys/kernel/debug
    mount -t tracefs tracefs /sys/kernel/tracing
    # END HACK :)
fi


if [ -z $1 ]; then
    echo "target client undefined"
    exit 1
fi

CONFIG_FILE=$(pwd)/config.toml
TARGET=$1
ERROR_MODEL_URL="$2"

get_config () {
    stoml $CONFIG_FILE $1
}

WORKING_DIR=$HOME
OUTPUT_DIR=$(get_config "output_dir")
CHAOS_ETH_DIR=$(get_config "chaos_eth_dir")
ERROR_MODELS="$WORKING_DIR/error_models.json"
wget -O $ERROR_MODELS $ERROR_MODEL_URL
PRE_SYNC_CMD=$(pwd)/synchronize-stop.sh

# spawn + sync wait
# { $PRE_SYNC_CMD $TARGET; }
echo "START" > ipc.dat

while true; do
    # start target
    TARGET_LOG="$OUTPUT_DIR/$TARGET-sync-$(date -Iseconds).log"
    TARGET_CMD=$(get_config "$TARGET.exec_cmd")
    DATA_DIR_PARAM=$(get_config "$TARGET.datadir_flag")=$WORKING_DIR/$(get_config "$TARGET.datadir")

    JWT_FLAG=$(get_config "$TARGET.jwt_flag")
    TARGET_JWT_FILE=$WORKING_DIR/$(get_config "$TARGET.jwt_path")
    if [ ! -z $JWT_FLAG ]; then
        JWT_PARAM="$JWT_FLAG=$TARGET_JWT_FILE"
    fi 

    { $TARGET_CMD $JWT_PARAM $DATA_DIR_PARAM &> $TARGET_LOG; } &
    TARGET_PPID=$!
    TARGET_GREP_STR=$TARGET_PPID.*$(get_config "$TARGET.grep_str")
    TARGET_PID=`ps axo pid,ppid,cmd | grep "$TARGET_GREP_STR" | awk '{print $1}'`
    sleep 60

    # start teku
    TEKU_LOG=$OUTPUT_DIR/teku-sync-$(date -Iseconds).log
    { teku --ee-endpoint=http://localhost:8551 --ee-jwt-secret-file=$TARGET_JWT_FILE --data-beacon-path=$WORKING_DIR/nvme/teku-data-dir/ &> $TEKU_LOG; } &
    TEKU_PPID=$!
    sleep 2
    TEKU_GREP_STR=$TEKU_PPID.*teku\\.home
    TEKU_PID=`ps axo pid,ppid,cmd | grep "$TEKU_GREP_STR" | awk '{print $1}'`

    #attach error injection
    sleep 10
    CHAOS_ETH_GREP_STR="[s]yscall_injector.py"
    
    { $SUDO python $CHAOS_ETH_DIR/syscall_injector.py --config $ERROR_MODELS -p $TARGET_PID &> $OUTPUT_DIR/chaos-$(date -Iseconds).log; } &
    CHAOS_ETH_PPID=$!
    CHAOS_ETH_GREP_STR=$CHAOS_ETH_PPID.*$CHAOS_ETH_GREP_STR
    CHAOS_ETH_PID=`ps axo pid,ppid,cmd | grep "$CHAOS_ETH_GREP_STR" | awk '{print $1}'`

    sleep 3

    # check that everything is still running

    TARGET_GREP="target"
    TEKU_GREP="teku"
    CHAOS_ETH_GREP="chaoseth"

    while [ ! -z "$TARGET_GREP" ] && [ ! -z "$TEKU_GREP" ] && [ ! -z "$CHAOS_ETH_GREP" ]
    do
        TARGET_GREP=`ps axo pid,ppid,cmd | grep "$TARGET_GREP_STR"`
        TEKU_GREP=`ps axo pid,ppid,cmd | grep "$TEKU_GREP_STR"`
        CHAOS_ETH_GREP=`ps axo pid,ppid,cmd | grep "$CHAOS_ETH_GREP_STR"`

        sleep 10
    done

    # if one crashed restart all

    kill -2 $TARGET_PID
    kill -2 $TEKU_PID
    $SUDO kill -2 $CHAOS_ETH_PID


    TARGET_GREP="target"
    TEKU_GREP="teku"
    CHAOS_ETH_GREP="chaoseth"

    while [ ! -z "$TARGET_GREP" ] || [ ! -z "$TEKU_GREP" ] || [ ! -z "$CHAOS_ETH_GREP" ]
    do
        sleep 1
        TARGET_GREP=`ps axo pid,ppid,cmd | grep "$TARGET_GREP_STR"`
        TEKU_GREP=`ps axo pid,ppid,cmd | grep "$TEKU_GREP_STR"`
        CHAOS_ETH_GREP=`ps axo pid,ppid,cmd | grep "$CHAOS_ETH_GREP_STR"`
    done

done
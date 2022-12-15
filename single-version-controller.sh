#!/bin/bash

TARGET=$1
ERROR_MODEL=$2

# run eth client!
{ ./single-version-fault-injection.sh $TARGET $ERROR_MODEL; } &
SUBSHELL=$!

# run workload!
echo "start random method workload"
./random-method-workload.sh
echo "end random method workload"

echo "start get block workload"
./get-block-workload.sh
echo "end get block workload"

kill -9 $SUBSHELL
# rm -rf $TARGET_DIR/*

# done
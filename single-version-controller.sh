#!/bin/bash

TARGET=$1
ERROR_MODEL=$2

# run eth client!
{ ./single-version-fault-injection.sh $TARGET $ERROR_MODEL; } &
SUBSHELL=$!

# run workload!
./random-method-workload.sh

kill -2 $SUBSHELL
# rm -rf $TARGET_DIR/*

# done
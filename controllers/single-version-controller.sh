#!/bin/bash

# copy data-dirs
SOURCE_DIR=/vol_src/
TARGET_DIR=/vol_dst

rsync -r --delete $SOURCE_DIR $TARGET_DIR

# run eth client!
./single-version-fault-injection.sh &

# run workload!
./random-method-workload.sh

# rm -rf $TARGET_DIR/*

# done
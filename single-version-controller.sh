#!/bin/bash

TARGET_DATA=/nvme

# with bucket 0 stopped!!!
echo "Stop client in bucket 0 before starting!"

sleep 20

# copy data-dirs
# SOURCE_DIR=$HOME/nvme/bucket0
# TARGET_DIR=$HOME/nvme/bucket$BUCKET

# rsync -r --delete $SOURCE_DIR $TARGET_DIR

# run eth client!
{./single-version-fault-injection.sh geth; } &

# run workload!
./random-method-workload.sh

# rm -rf $TARGET_DIR/*

# done
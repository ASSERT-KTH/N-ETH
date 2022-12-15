#!/bin/bash

echo "WAIT" > ipc.dat

VAL=$(cat ipc.dat)

while [[ $VAL = "WAIT" ]]
do
    VAL=$(cat ipc.dat)
    sleep 1m
done

echo "Starting requests..."
go run requests-get-block.go &> /output/requests-block.log

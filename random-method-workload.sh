#!/bin/bash

VAL=$(cat ipc.dat)

while [[ $VAL = "WAIT" ]]
do
    VAL=$(cat ipc.dat)
    sleep 1m
done

echo "Starting requests..."
go run requests-random.go &> /output/requests-random.log

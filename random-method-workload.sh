#!/bin/bash

echo "WAIT" > ipc.dat

VAL=$(cat ipc.dat)

while [[ $VAL = "WAIT" ]]
do
    VAL=$(cat ipc.dat)
    sleep 1m
done

go run requests.go > requests.log

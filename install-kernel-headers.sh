#!/bin/bash

TARGET=$1

docker build --build-arg target=$TARGET -f kernel-headers.dockerfile -t javierron/neth:$TARGET-kernel .
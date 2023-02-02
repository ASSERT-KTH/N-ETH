#!/bin/sh

trap shutdown SIGINT SIGQUIT SIGTERM

shutdown()
{
    exit
}

./proxy adaptive
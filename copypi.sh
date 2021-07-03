#!/bin/sh
if [ $# -lt 1 ]; then
    echo arg error: $*
    exit 1
fi

scp -p pirunner pi@$1:/home/pi/radigo/

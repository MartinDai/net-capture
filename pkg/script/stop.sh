#!/bin/bash

set -e

STR=`ps -ef | grep "net-capture" | awk  -F ' '  '{print $2}'`
if [ ! -z "${STR}" ]; then
    kill -9 ${STR} > /dev/null 2>&1
    sleep 1
    echo "Stop net-capture successful"
else
  echo "no net-capture process running"
fi

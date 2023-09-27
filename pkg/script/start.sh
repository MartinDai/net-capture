#!/bin/bash

set -e

SHELL_HOME="$(cd "$(dirname "$0")" && pwd)"

nohup ${SHELL_HOME}/net-capture --config-file ${SHELL_HOME}/config.yml > ${SHELL_HOME}/capture.log 2>&1 &

echo "Agent started, please go to the directory ${SHELL_HOME}/capture.log to view the logs."

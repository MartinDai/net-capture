@echo off
start /B net-capture.exe -config-file=./config.yml > capture.log 2>&1

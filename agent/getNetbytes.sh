#!/bin/bash
DATE=`date --utc`
RX=`cat /proc/net/dev | grep em1 |awk '{print $2}'`
TX=`cat /proc/net/dev | grep em1 |awk '{print $10}'`
echo "$RX;$TX"
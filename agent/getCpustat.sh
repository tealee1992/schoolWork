#!/bin/bash
idle=`sar -u 1 1 | sed -n '4p'|awk '{  i=NF; print $i }'`
echo "$idle"
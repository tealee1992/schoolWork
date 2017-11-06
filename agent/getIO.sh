#!/bin/bash
echo `iostat -x 1 2|grep sda |awk '{print $14}'|sed -n '2p'`
#!/bin/bash

masterIP=$1
conID=$2
echo $masterIP
echo $conID
//注意，这里的docker命令不能使用-ti，不然会出现 “the input device is not a TTY”的错误
cmd="docker -H $masterIP:3375 exec -i $conID bash -c \"export DISPLAY=:1 && bash /tempfiles/eclipseReload.sh\""
# cmd="docker -H $materIP:3375 exec -ti $conID bash -c \"ls\""
# cmd="docker -H 11.0.0.172:3375 exec -ti de9b07e93d63 bash -c \"ls\""
eval $cmd

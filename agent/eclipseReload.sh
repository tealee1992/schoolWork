#!/bin/bash
# this script is for processing stuff before stop container
# need excute "export DISPLAY=:1" first in the docker exec command
WID=`xdotool search --name "Eclipse Platform" | head -1`
if [ -n "$WID" ]
then
	xdotool windowactivate $WID
	xdotool windowfocus $WID
	xdotool key ctrl+S
	# xdotool windowkill $WID
	xdotool key --window $WID Return
	xdotool windowkill $WID
else
	echo "eclipse is not running"
fi	


#!/bin/bash
# this script is for processing stuff before stop container
WID=`xdotool search --name "Eclipse Platform" | head -1`
if [ -n "$WID" ]
then
	xdotool windowactivate $WID
	xdotool windowfocus $WID
	xdotool key ctrl+S
	xdotool windowkill $WID
else
	echo"eclipse is not running"
fi	


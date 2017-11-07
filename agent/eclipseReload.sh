#!/bin/bash
# this script is for processing stuff before stop container
WID=`xdotool search --name "Eclipse Platform" | head -1`
if [-z WID]; then
	echo"eclipse is not running"
else
	xdotool windowactivate $WID
	xdotool windowfocus $WID
	xdotool key ctrl+S
	xdotool windowkill $WID
fi	


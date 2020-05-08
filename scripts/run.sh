#!/bin/sh
rmmod ./pi-joystick/js-audio.ko
insmod ./pi-joystick/js-audio.ko
nohup ./gamepad-agent > /dev/null &

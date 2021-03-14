#!/usr/bin/sh

log=build.log

while inotifywait -qq -e create -e modify .; do
    echo -n "$(date +%F:%T) " >> $log
    make >> $log
done

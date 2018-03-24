#!/usr/bin/env bash
for k in $( seq 1 4 )
do
    pkill -9 -o "node${k}" &
done
pkill -9 -o node36
pkill -9 -o node37
pkill -9 -o node38
wait
exit
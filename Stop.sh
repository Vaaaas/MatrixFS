#!/usr/bin/env bash
for k in $( seq 0 3 )
do
    pkill -9 -o "node${k}" &
done
pkill -9 -o node40
pkill -9 -o node41
wait
exit
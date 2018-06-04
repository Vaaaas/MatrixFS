#!/usr/bin/env bash
for k in $( seq 1 6 )
do
    pkill -9 -o "node${k}" &
done
wait
exit
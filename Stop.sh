#!/usr/bin/env bash
for k in $( seq 0 25 )
do
    pkill -9 "node${k}" &
done
wait
exit
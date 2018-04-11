#!/usr/bin/env bash
for k in $( seq 0 5 )
do
    pkill -9 -o "node${k}" &
done
wait
exit
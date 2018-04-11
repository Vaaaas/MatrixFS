#!/usr/bin/env bash
port=9090
for k in $( seq 0 7 )
do
    docker run -p ${port}:${port} --name Node${k} vaaaas/matrixfs_node:1.0.0 -master="111.230.208.23:8080" -node="222.18.167.199:${port}" &
    let port++
done
wait
exit

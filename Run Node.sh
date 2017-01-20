#!/usr/bin/env bash
cd node;
port=9090
for k in $( seq 0 25 )
do
    go build -o "/Users/vaaaas/OneDrive/Software/Go/src/github.com/Vaaaas/MatrixFS/node/Node${k}" "/Users/vaaaas/OneDrive/Software/Go/src/github.com/Vaaaas/MatrixFS/node/node.go";
    ./"node${k}" -stpath="./storage${k}" -log_dir="./log${k}" -node="127.0.0.1:${port}" &
    let port++
done
wait
exit

#go build -o node\Node.exe node\node.go;
#cd node;
#node -log_dir="./log" -master="192.168.199.201:8080" -node="192.168.199.220:9090"
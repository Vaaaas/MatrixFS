#!/usr/bin/env bash
cd node;
go build -o "/Users/vaaaas/OneDrive/Software/Go/src/github.com/Vaaaas/MatrixFS/node/Node 01";
./"node 01" -stpath="./storage01" -log_dir="./log01" -node="127.0.0.1:9091" &
go build -o "/Users/vaaaas/OneDrive/Software/Go/src/github.com/Vaaaas/MatrixFS/node/Node 02";
./"node 02" -stpath="./storage02" -log_dir="./log02" -node="127.0.0.1:9092" &
go build -o "/Users/vaaaas/OneDrive/Software/Go/src/github.com/Vaaaas/MatrixFS/node/Node 03";
./"node 03" -stpath="./storage03" -log_dir="./log03" -node="127.0.0.1:9093" &
go build -o "/Users/vaaaas/OneDrive/Software/Go/src/github.com/Vaaaas/MatrixFS/node/Node 04";
./"node 04" -stpath="./storage04" -log_dir="./log04" -node="127.0.0.1:9094" &
go build -o "/Users/vaaaas/OneDrive/Software/Go/src/github.com/Vaaaas/MatrixFS/node/Node 05";
./"node 05" -stpath="./storage05" -log_dir="./log05" -node="127.0.0.1:9095" &
wait
exit
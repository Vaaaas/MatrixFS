#!/usr/bin/env bash
go get github.com/golang/glog
mkdir -p $HOME/MatrixFS/node;
cd $HOME/MatrixFS/node;
port=9090
#2 Faults 2 Rows
for k in $( seq 1 8 )
do
    go build -i -o "$HOME/MatrixFS/node/node${k}" github.com/Vaaaas/MatrixFS/node;
    ./node${k} -stpath="./storage${k}" -log_dir="./log${k}" -node="127.0.0.1:${port}"&
    let port++
done
wait
exit

#!/usr/bin/env bash
mkdir -p $HOME/MatrixFS/node;
cd $HOME/MatrixFS/node;
port=9090
for k in $( seq 0 85 )
do
    go build -i -o "$HOME/MatrixFS/node/node${k}" github.com/Vaaaas/MatrixFS/node;
    ./node${k} -stpath="./storage${k}" -log_dir="./log${k}" -master="127.0.0.1:8080" -node="127.0.0.1:${port}"&
    let port++
done
wait
exit

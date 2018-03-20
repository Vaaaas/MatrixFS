#!/usr/bin/env bash
mkdir -p $HOME/MatrixFS/node
#go build -i -o $HOME/MatrixFS/node/node github.com/Vaaaas/MatrixFS/node;
cd $HOME/MatrixFS/node;
port=9090
for k in $( seq 0 14 )
do
    ./node -stpath="./storage${k}" -log_dir="./log${k}" -node="127.0.0.1:${port}" &
    let port++
done
wait
exit
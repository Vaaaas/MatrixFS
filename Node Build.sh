#!/usr/bin/env bash
mkdir -p $HOME/MatrixFS/node
go build -i -o $HOME/MatrixFS/node/node github.com/Vaaaas/MatrixFS/node;
exit
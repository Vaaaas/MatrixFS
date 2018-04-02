#!/usr/bin/env bash
go get github.com/golang/glog
go build -i -o $HOME/MatrixFS/master/master github.com/Vaaaas/MatrixFS;
cd $HOME/MatrixFS/master;
./master -log_dir=./log -alsologtostderr=true
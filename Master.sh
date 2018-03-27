#!/usr/bin/env bash
go get github.com/golang/glog
mkdir -p $HOME/MatrixFS/master/js/ && cp -rf $GOPATH/src/github.com/Vaaaas/MatrixFS/server/js $HOME/MatrixFS/master/;
mkdir -p $HOME/MatrixFS/master/css/ && cp -rf $GOPATH/src/github.com/Vaaaas/MatrixFS/server/css $HOME/MatrixFS/master/;
mkdir -p $HOME/MatrixFS/master/view/ && cp -rf $GOPATH/src/github.com/Vaaaas/MatrixFS/server/view $HOME/MatrixFS/master/;
mkdir -p $HOME/MatrixFS/master/image/ && cp -rf $GOPATH/src/github.com/Vaaaas/MatrixFS/server/image $HOME/MatrixFS/master/;
cp -f $GOPATH/src/github.com/Vaaaas/MatrixFS/server/favicon.ico $HOME/MatrixFS/master/favicon.ico;
go build -i -o $HOME/MatrixFS/master/master github.com/Vaaaas/MatrixFS/server;
cd $HOME/MatrixFS/master;
./master -log_dir=./log -alsologtostderr=true
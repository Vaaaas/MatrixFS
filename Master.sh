#!/usr/bin/env bash
mkdir -p $HOME/MatrixFS/master/js/ && cp -rf $GOPATH/src/github.com/Vaaaas/MatrixFS/master/js $HOME/MatrixFS/master/;
mkdir -p $HOME/MatrixFS/master/css/ && cp -rf $GOPATH/src/github.com/Vaaaas/MatrixFS/master/css $HOME/MatrixFS/master/;
mkdir -p $HOME/MatrixFS/master/view/ && cp -rf $GOPATH/src/github.com/Vaaaas/MatrixFS/master/view $HOME/MatrixFS/master/;
mkdir -p $HOME/MatrixFS/master/image/ && cp -rf $GOPATH/src/github.com/Vaaaas/MatrixFS/master/image $HOME/MatrixFS/master/;
cp -f $GOPATH/src/github.com/Vaaaas/MatrixFS/master/favicon.ico $HOME/MatrixFS/master/favicon.ico;
go build -i -o $HOME/MatrixFS/master/master github.com/Vaaaas/MatrixFS/master;
cd $HOME/MatrixFS/master;
./master -log_dir=./log -alsologtostderr=true
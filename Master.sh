#!/usr/bin/env bash
mkdir -p $HOME/MatrixFS/master/js/ && cp -rf $GOPATH/src/github.com/Vaaaas/MatrixFS/matrix/js $HOME/MatrixFS/master/;
mkdir -p $HOME/MatrixFS/master/pic/ && cp -rf $GOPATH/src/github.com/Vaaaas/MatrixFS/matrix/pic/ $HOME/MatrixFS/master/;
mkdir -p $HOME/MatrixFS/master/view/ && cp -rf $GOPATH/src/github.com/Vaaaas/MatrixFS/matrix/view/ $HOME/MatrixFS/master/;
cp -f $GOPATH/src/github.com/Vaaaas/MatrixFS/matrix/favicon.ico $HOME/MatrixFS/master/favicon.ico;
go build -i -o $HOME/MatrixFS/master/master github.com/Vaaaas/MatrixFS/matrix;
cd $HOME/MatrixFS/master;
./master -log_dir=./log -alsologtostderr=true
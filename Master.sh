#!/usr/bin/env bash
mkdir -p $HOME/MatrixFS/master/js/ && cp  -rf $GOPATH/src/github.com/Vaaaas/MatrixFS/matrix/js/ $HOME/MatrixFS/master/js/
mkdir -p $HOME/MatrixFS/master/pic/ && cp  -rf $GOPATH/src/github.com/Vaaaas/MatrixFS/matrix/pic/ $HOME/MatrixFS/master/pic/
mkdir -p $HOME/MatrixFS/master/view/ && cp  -rf $GOPATH/src/github.com/Vaaaas/MatrixFS/matrix/view/ $HOME/MatrixFS/master/view/
cp -f $GOPATH/src/github.com/Vaaaas/MatrixFS/matrix/favicon.ico $HOME/MatrixFS/master/favicon.ico
go build -i -o $HOME/MatrixFS/master/master github.com/Vaaaas/MatrixFS/matrix;
cd $HOME/MatrixFS/master;
mkdir log;
mkdir temp;
./master -log_dir=./log -alsologtostderr=true

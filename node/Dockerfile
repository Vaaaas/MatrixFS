FROM golang:latest

RUN go get github.com/golang/glog
RUN go get github.com/Vaaaas/MatrixFS/node

COPY . /node
WORKDIR /node
RUN go build .

EXPOSE 9090

ENTRYPOINT ["./node","-stpath=./storage","-log_dir=./log"]
FROM golang:latest

RUN go get github.com/golang/glog
RUN go get github.com/Vaaaas/MatrixFS

COPY . /master
WORKDIR /master
RUN go build .

EXPOSE 8080

ENTRYPOINT ["./master","-log_dir=./log","-alsologtostderr=true"]
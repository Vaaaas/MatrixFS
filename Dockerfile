FROM golang:latest AS build-env

RUN go get github.com/golang/glog \
    && go get github.com/Vaaaas/MatrixFS

COPY . /master
WORKDIR /master
RUN ls

RUN CGO_ENABLED=0 go build -o master .

FROM alpine:latest
RUN apk add -U tzdata
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime
COPY --from=build-env /master /master
RUN cd master/server \
    && ls
RUN chmod +x /master
WORKDIR /master
EXPOSE 8080
ENTRYPOINT ["./master","-log_dir=./log","-alsologtostderr=true"]
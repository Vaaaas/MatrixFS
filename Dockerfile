FROM golang:latest AS build-env

RUN go get github.com/golang/glog \
    && go get github.com/Vaaaas/MatrixFS

COPY . /master
WORKDIR /master

RUN CGO_ENABLED=0 go build -o master .

FROM alpine:latest
COPY --from=build-env /master/master .
RUN chmod +x master
EXPOSE 8080
ENTRYPOINT ["./master","-log_dir=./log","-alsologtostderr=true"]
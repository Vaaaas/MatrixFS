FROM golang:latest AS build-env

RUN go get github.com/golang/glog \
    && go get github.com/Vaaaas/MatrixFS/node

COPY . /node
WORKDIR /node
RUN CGO_ENABLED=0 go build -o node .

FROM alpine:latest
COPY --from=build-env /node/node .
RUN chmod +x node

EXPOSE 9090

ENTRYPOINT ["./node","-stpath=./storage","-log_dir=./log"]
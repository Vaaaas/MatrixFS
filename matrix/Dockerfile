FROM golang
WORKDIR /master
COPY /js /js
COPY /pic /pic
COPY /view /view
COPY favicon.ico favicon.ico
RUN go get github.com/Vaaaas/MatrixFS/matrix
EXPOSE 8080
VOLUME /temp
VOLUME /log
ENTRYPOINT ["matrix"]

FROM golang:1.15 AS build-env

WORKDIR /go/bin/go-icn

COPY . .

WORKDIR /go/bin/go-icn/client

RUN go mod download
RUN go get github.com/githubnemo/CompileDaemon

ENTRYPOINT CompileDaemon --build="go build" --command=./client
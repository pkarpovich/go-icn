FROM golang:1.15 AS build-env

WORKDIR /go/bin/go-icn

RUN mkdir /go/bin/go-icn/utils
RUN mkdir /go/bin/go-icn/content-provider

COPY ./utils ./utils

WORKDIR /go/bin/go-icn/content-provider
COPY ./content-provider/go.mod .
COPY ./content-provider/go.sum .
RUN go mod download
COPY ./content-provider .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

FROM scratch
EXPOSE 8080

WORKDIR /app
COPY --from=build-env /go/bin/go-icn/content-provider/content-provider ./content-provider

ENTRYPOINT ["/app/content-provider"]
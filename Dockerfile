FROM golang:1.11.5

LABEL maintainer="bik" version="1.0"

WORKDIR /go/src/github.com/dzeckelev/adviser

COPY . .

RUN go install -v ./...
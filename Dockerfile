FROM golang:1.4.2
MAINTAINER peter.edge@gmail.com

RUN mkdir -p /go/src/github.com/peter-edge/go-gen-enumtype
ADD . /go/src/github.com/peter-edge/go-gen-enumtype/
WORKDIR /go/src/github.com/peter-edge/go-gen-enumtype

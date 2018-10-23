FROM golang:1.11

#RUN go get -u google.golang.org/grpc
RUN go get github.com/golang/protobuf/protoc-gen-go
RUN go get github.com/golang/mock/gomock
RUN go install github.com/golang/mock/mockgen

RUN apt-get update && apt-get install unzip

COPY install.sh /install.sh
COPY entrypoint.sh /entrypoint.sh

RUN /install.sh

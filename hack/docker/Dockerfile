FROM golang:1.11

RUN apt-get update && apt-get install -y unzip && apt-get upgrade -y

# Protobuf install specific version
RUN go get github.com/golang/protobuf/...
RUN rm -f /go/bin/protoc-gen-go
WORKDIR /go/src/github.com/golang/protobuf/
RUN git checkout tags/v1.2.0 && go install /go/src/github.com/golang/protobuf/protoc-gen-go/

RUN go get github.com/golang/mock/gomock
RUN go install github.com/golang/mock/mockgen

RUN go get golang.org/x/tools/cmd/cover
RUN go get github.com/mattn/goveralls

WORKDIR /

COPY install.sh /install.sh
COPY entrypoint.sh /entrypoint.sh

RUN /install.sh

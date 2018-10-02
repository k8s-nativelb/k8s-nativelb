# Build the manager binary
FROM golang:1.10.3 as builder

# Copy in the go src
WORKDIR /go/src/github.com/k8s-nativelb
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o nativelb-controller github.com/k8s-nativelb/cmd/nativelb-controller

# Copy the controller-manager into a thin image
FROM centos:latest
WORKDIR /root/
COPY --from=builder /go/src/github.com/k8s-nativelb/nativelb-controller .
ENTRYPOINT ["./nativelb-controller"]

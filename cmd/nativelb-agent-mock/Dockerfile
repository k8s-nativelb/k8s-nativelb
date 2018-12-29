# Build the manager binary
FROM golang:1.11.0 as builder

# Copy in the go src
WORKDIR /go/src/github.com/k8s-nativelb
COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o nativelb-agent-mock github.com/k8s-nativelb/cmd/nativelb-agent-mock

# Copy the agent into a thin image
FROM fedora:29
WORKDIR /root/
COPY --from=builder /go/src/github.com/k8s-nativelb/nativelb-agent-mock .
ENTRYPOINT ["./nativelb-agent-mock"]

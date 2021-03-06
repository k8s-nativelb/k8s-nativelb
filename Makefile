# Image URL to use all building/pushing image targets
REGISTRY ?= quay.io
IMG_CONTROLLER ?= k8s-nativelb/nativelb-controller
IMG_API ?= k8s-nativelb/nativelb-api
IMG_AGENT ?= k8s-nativelb/nativelb-agent
IMG_AGENT_MOCK ?= k8s-nativelb/nativelb-agent-mock

IMAGE_TAG ?= latest

all: docker-make

# Build manager binary
#manager: generate fmt vet
#	go build -o bin/manager github.com/k8s-external-lb/native-lb-controller/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
#run: generate fmt vet
#	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install:
	kubectl apply -f config/rbac
	kubectl apply -f config/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: install
	kubectl create ns nativelb
	kubectl apply -f config/release

# Generate manifests e.g. CRD, RBAC etc.
# TODO: need to fix the CRD generator remove the status section. then return the command to all
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all

crd:
	./hack/crd.sh

rbac:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go rbac
	mv config/rbac/* config/base/rbac/
	rm -rf config/rbac

# Generate code
generate:
	protoc -I. proto/native-lb.proto --go_out=plugins=grpc:.
	cp proto/native-lb.pb.go pkg/proto/proto.pb.go
	mockgen -source pkg/proto/proto.pb.go -package=proto -destination=pkg/proto/generated_mock_proto.go
	go generate ./pkg/... ./cmd/...
	./hack/crd.sh
	kustomize build config/default/release/ > config/release/k8s-nativelb.yaml
	kustomize build config/default/test/ > config/test/k8s-nativelb.yaml

vet:
	go vet ./pkg/... ./cmd/... ./tests/...

fmt:
	go fmt ./pkg/... ./cmd/... ./tests/...

# Run tests
test:
	go test -v -race ./pkg/... ./cmd/... -coverprofile cover.out 

functest:
	go test -v -race ./tests/.

goveralls:
	./hack/goveralls.sh

#### Docker section ###
docker-make:
	./hack/run.sh generate fmt vet

docker-generate:
	./hack/run.sh generate fmt vet

docker-goveralls: docker-test
	./hack/run.sh goveralls

# Test Inside a docker
docker-test:
	./hack/run.sh test

docker-functest:
	./hack/run.sh functest

# Build the docker image
docker-build:
	docker build -f./cmd/nativelb-controller/Dockerfile -t ${REGISTRY}/${IMG_CONTROLLER}:${IMAGE_TAG} .
	docker build -f./cmd/nativelb-agent/Dockerfile -t ${REGISTRY}/${IMG_AGENT}:${IMAGE_TAG} .
	docker build -f./cmd/nativelb-agent-mock/Dockerfile -t ${REGISTRY}/${IMG_AGENT_MOCK}:${IMAGE_TAG} .

# Push the docker image
docker-push:
	docker push ${REGISTRY}/${IMG_CONTROLLER}:${IMAGE_TAG}
	docker push ${REGISTRY}/${IMG_AGENT}:${IMAGE_TAG}
	docker push ${REGISTRY}/${IMG_AGENT_MOCK}:${IMAGE_TAG}

cluster-up:
	./cluster/up.sh

cluster-down:
	./cluster/down.sh

cluster-sync:
	./cluster/sync.sh


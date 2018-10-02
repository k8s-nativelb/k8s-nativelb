# Image URL to use all building/pushing image targets
IMG_CONTROLLER ?= quay.io/k8s-nativelb/nativelb-controller
IMG_API ?= quay.io/k8s-nativelb/nativelb-api
IMG_AGENT ?= quay.io/k8s-nativelb/nativelb-agent
IMAGE_TAG ?= latest


all: docker-test docker-build deploy

# Run tests
test: generate fmt vet #manifests
	go test ./pkg/... ./cmd/... -coverprofile cover.out

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
	kubectl apply -f config/samples
	#kubectl apply -f config/manager

# Generate manifests e.g. CRD, RBAC etc.
# TODO: need to fix the CRD generator remove the status section. then return the command to all
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all

# TODO: need to fix this
crd:
	echo "Need to update the crd manualy (remove status and things other then Proterties"
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go crd

rbac:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go rbac

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
	go generate ./pkg/... ./cmd/...

functest:
	go test ./tests/.

# Test Inside a docker
docker-test:
	./test.sh

docker-functest:
	./func-test.sh

# Build the docker image
docker-build: docker-test
	docker build -f./cmd/nativelb-controller/Dockerfile -t ${IMG_CONTROLLER}:${IMAGE_TAG} .

# Push the docker image
docker-push: docker-build
	docker push ${IMG_CONTROLLER}:${IMAGE_TAG}

#publish:
#	docker build . -t ${IMG}
#	docker push ${IMG}

#oc-cluster-up:
#	oc cluster up --base-dir=/opt/openshift/

#oc-cluster-down:
#	oc cluster down

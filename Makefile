# Image URL to use all building/pushing image targets
IMG_CONTROLLER ?= quay.io/k8s-nativelb/k8s-nativelb-controller:latest
IMG_API ?= quay.io/k8s-nativelb/k8s-nativelb-api:latest
IMG_AGENT ?= quay.io/k8s-nativelb/k8s-nativelb-agent:latest


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

print:
	./hack/travis.sh

# Test Inside a docker
#docker-test:
#	./build-test.sh

# Build the docker image
#docker-build: docker-test
#	docker build . -t ${IMG}

# Push the docker image
#docker-push: docker-build
#	docker push ${IMG}

#publish:
#	docker build . -t ${IMG}
#	docker push ${IMG}

#oc-cluster-up:
#	oc cluster up --base-dir=/opt/openshift/

#oc-cluster-down:
#	oc cluster down

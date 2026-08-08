# Default image tag and Helm release configuration
IMG ?= ghcr.io/muhmmadayan/fake-network-operator:0.1.0
HELM_RELEASE ?= fake-network-operator
HELM_NAMESPACE ?= fake-network-operator
HELM_CHART ?= deploy/fake-network-operator

# Docker runner helper pattern for executing Go tools inside container
DOCKER_GO_RUN = docker run --rm -v $(shell pwd):/workspace -w /workspace golang:1.23-alpine

.PHONY: all docker-build docker-buildx docker-push helm-lint helm-template helm-install helm-uninstall generate fmt vet test clean

# Default target
all: docker-build

# Build the container image containing all operator binaries
docker-build:
	docker build -t $(IMG) .

# Build multi-arch container image
docker-buildx:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(IMG) --push .

# Push the container image to registry
docker-push:
	docker push $(IMG)

# Remove locally built container images
clean:
	-docker rmi $(IMG)

# Helm chart linting
helm-lint:
	helm lint $(HELM_CHART)

# Helm chart rendering dry-run for debugging
helm-template:
	helm template $(HELM_RELEASE) $(HELM_CHART)

# Install or upgrade Helm release on Kubernetes cluster
helm-install:
	helm upgrade --install $(HELM_RELEASE) $(HELM_CHART) \
		--namespace $(HELM_NAMESPACE) \
		--create-namespace

# Uninstall Helm release
helm-uninstall:
	helm uninstall $(HELM_RELEASE) --namespace $(HELM_NAMESPACE)

# Generate CRD manifests and DeepCopy methods using controller-gen inside Docker
generate:
	./hack/generate.sh

# Run go fmt inside Docker container
fmt:
	$(DOCKER_GO_RUN) go fmt ./...

# Run go vet inside Docker container
vet:
	$(DOCKER_GO_RUN) go vet ./...

# Run unit tests inside Docker container
test:
	$(DOCKER_GO_RUN) go test -v ./...

# Run unit tests with coverage report
coverage:
	$(DOCKER_GO_RUN) go test -v -coverprofile=coverage.out ./...
	$(DOCKER_GO_RUN) go tool cover -html=coverage.out -o coverage.html


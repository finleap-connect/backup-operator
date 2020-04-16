# Directory, where all required tools are located (absolute path required)
TOOLS_DIR ?= $(shell cd tools && pwd)

# Prerequisite tools
GO ?= go
DOCKER ?= docker
KUBECTL ?= kubectl

# Managed by this project
GINKGO ?= $(TOOLS_DIR)/ginkgo
LINTER ?= $(TOOLS_DIR)/golangci-lint
KIND ?= $(TOOLS_DIR)/kind
HELM3 ?= $(TOOLS_DIR)/helm3
CONTROLLER_GEN ?= $(TOOLS_DIR)/controller-gen
KUSTOMIZE ?= $(TOOLS_DIR)/kustomize
KUBEBUILDER ?= $(TOOLS_DIR)/kubebuilder
KUBEBUILDER_ASSETS ?= $(TOOLS_DIR)

MANAGER_BIN ?= bin/manager
MANAGER_IMG ?= kubismio/manager:latest

WORKER_BIN ?= bin/worker
WORKER_IMG ?= kubismio/worker:latest

export

.PHONY: all test lint fmt vet install uninstall deploy manifests docker-build docker-push tools

all: $(MANAGER_BIN) $(WORKER_BIN) tools

$(MANAGER_BIN): generate fmt vet
	$(GO) build -o $(MANAGER_BIN) ./cmd/manager/main.go

$(WORKER_BIN): generate fmt vet
	$(GO) build -o $(WORKER_BIN) ./cmd/worker/...

test: generate fmt vet manifests $(GINKGO) $(KUBEBUILDER)
	$(GINKGO) -r -v -coverprofile cover.out pkg test/interaction

integration: generate manifests docker-build $(GINKGO) $(KIND)
	$(GINKGO) -r -v test/integration

lint: $(LINTER)
	$(GO) mod verify
	$(LINTER) run -v --no-config --deadline=5m

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

# Install CRDs into a cluster
install: manifests $(KUSTOMIZE)
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests $(KUSTOMIZE)
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests $(KUSTOMIZE)
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(MANAGER_IMG)
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) crd:trivialVersions=false rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Generate code
generate: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

docker-build: docker-build-manager docker-build-worker

docker-build-manager:
	$(DOCKER) build --target manager . -t $(MANAGER_IMG)

docker-build-worker:
	$(DOCKER) build --target worker . -t $(WORKER_IMG)

docker-push: docker-push-manager docker-push-worker

docker-push-manager:
	$(DOCKER) push $(MANAGER_IMG)

docker-push-worker:
	$(DOCKER) push $(WORKER_IMG)

# Phony target to install all required tools into ${TOOLS_DIR}
tools: $(TOOLS_DIR)/kind $(TOOLS_DIR)/ginkgo $(TOOLS_DIR)/controller-gen $(TOOLS_DIR)/kustomize $(TOOLS_DIR)/golangci-lint $(TOOLS_DIR)/kubebuilder $(TOOLS_DIR)/helm3

$(TOOLS_DIR)/kind:
	$(shell $(TOOLS_DIR)/goget-wrapper sigs.k8s.io/kind@v0.7.0)

$(TOOLS_DIR)/ginkgo:
	$(shell $(TOOLS_DIR)/goget-wrapper github.com/onsi/ginkgo/ginkgo@v1.12.0)

$(TOOLS_DIR)/controller-gen:
	$(shell $(TOOLS_DIR)/goget-wrapper sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5)

$(TOOLS_DIR)/kustomize:
	$(shell $(TOOLS_DIR)/goget-wrapper sigs.k8s.io/kustomize/kustomize/v3@v3.5.3)

$(TOOLS_DIR)/golangci-lint:
	$(shell curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_DIR) v1.24.0)

$(TOOLS_DIR)/kubebuilder $(TOOLS_DIR)/kubectl $(TOOLS_DIR)/kube-apiserver $(TOOLS_DIR)/etcd:
	$(shell $(TOOLS_DIR)/kubebuilder-install)

$(TOOLS_DIR)/helm3:
	$(shell $(TOOLS_DIR)/helm3-install)


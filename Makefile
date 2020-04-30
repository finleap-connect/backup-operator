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
GOVERALLS ?= $(TOOLS_DIR)/goveralls
GOVER ?= $(TOOLS_DIR)/gover
HELM3 ?= $(TOOLS_DIR)/helm3
CONTROLLER_GEN ?= $(TOOLS_DIR)/controller-gen
KUSTOMIZE ?= $(TOOLS_DIR)/kustomize
KUBEBUILDER ?= $(TOOLS_DIR)/kubebuilder
KUBEBUILDER_ASSETS ?= $(TOOLS_DIR)

MANAGER_BIN ?= bin/manager
WORKER_BIN ?= bin/worker

DOCKER_TAG ?= latest
DOCKER_IMG ?= kubismio/backup-operator:$(DOCKER_TAG)

KIND_CLUSTER ?= test
KIND_IMAGE ?= kindest/node:v1.16.4

HELM_CHART_DIR ?= charts/backup-operator

# Options to control behavior
TEST_E2E ?=

export

.PHONY: all test lint fmt vet install uninstall deploy manifests docker-build docker-push tools docker-is-running kind-create kind-delete kind-is-running

all: $(MANAGER_BIN) $(WORKER_BIN) tools

$(MANAGER_BIN): generate fmt vet
	$(GO) build -o $(MANAGER_BIN) ./cmd/manager/main.go

$(WORKER_BIN): generate fmt vet
	$(GO) build -o $(WORKER_BIN) ./cmd/worker/...

test: generate fmt vet manifests docker-is-running kind-is-running check-e2e $(GINKGO) $(KUBEBUILDER)
	$(GINKGO) -r -v -cover pkg

test-%: generate fmt vet manifests docker-is-running kind-is-running check-e2e $(GINKGO) $(KUBEBUILDER)
	$(GINKGO) -r -v -cover pkg/$*

check-e2e:
ifdef TEST_E2E
	$(MAKE) docker-build
endif

coverage: $(GOVERALLS) $(GOVER)
	$(GOVER)
	$(GOVERALLS) -coverprofile=gover.coverprofile -service=travis-ci -repotoken $(COVERALLS_TOKEN)

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
	$(KUSTOMIZE) build config/crd > $(HELM_CHART_DIR)/files/crds.yaml
	$(KUSTOMIZE) build config/rbac-templates > $(HELM_CHART_DIR)/templates/rbac.yaml

# Generate code
generate: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

docker-build:
	$(DOCKER) build . -t $(DOCKER_IMG)

docker-push:
	$(DOCKER) push $(DOCKER_IMG)

docker-is-running:
	@echo "Checking if docker is running..."
	@{ \
	set -e; \
	$(DOCKER) version > /dev/null; \
	}

kind-create: $(KIND)
	$(KIND) create cluster --image $(KIND_IMAGE) --name $(KIND_CLUSTER) --wait 5m

kind-is-running: $(KIND)
	@echo "Checking if kind cluster with name '$(KIND_CLUSTER)' is running..."
	@echo "(e.g. create cluster via 'make kind-create')"
	@{ \
	set -e; \
	$(KIND) get kubeconfig --name $(KIND_CLUSTER) > /dev/null; \
	}

kind-delete: $(KIND)
	$(KIND) delete cluster --name $(KIND_CLUSTER)


# Phony target to install all required tools into ${TOOLS_DIR}
tools: $(TOOLS_DIR)/kind $(TOOLS_DIR)/ginkgo $(TOOLS_DIR)/controller-gen $(TOOLS_DIR)/kustomize $(TOOLS_DIR)/golangci-lint $(TOOLS_DIR)/kubebuilder $(TOOLS_DIR)/helm3 $(TOOLS_DIR)/goveralls $(TOOLS_DIR)/gover

$(TOOLS_DIR)/kind:
	$(shell $(TOOLS_DIR)/goget-wrapper sigs.k8s.io/kind@v0.7.0)

$(TOOLS_DIR)/ginkgo:
	$(shell $(TOOLS_DIR)/goget-wrapper github.com/onsi/ginkgo/ginkgo@v1.12.0)

$(TOOLS_DIR)/controller-gen:
	$(shell $(TOOLS_DIR)/goget-wrapper sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5)

$(TOOLS_DIR)/kustomize:
	$(shell $(TOOLS_DIR)/goget-wrapper sigs.k8s.io/kustomize/kustomize/v3@v3.5.3)

$(TOOLS_DIR)/golangci-lint:
	$(shell curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_DIR) v1.25.0)

$(TOOLS_DIR)/kubebuilder $(TOOLS_DIR)/kubectl $(TOOLS_DIR)/kube-apiserver $(TOOLS_DIR)/etcd:
	$(shell $(TOOLS_DIR)/kubebuilder-install)

$(TOOLS_DIR)/helm3:
	$(shell $(TOOLS_DIR)/helm3-install)

$(TOOLS_DIR)/goveralls:
	$(shell $(TOOLS_DIR)/goget-wrapper github.com/mattn/goveralls@v0.0.5)

$(TOOLS_DIR)/gover:
	$(shell $(TOOLS_DIR)/goget-wrapper github.com/modocache/gover)


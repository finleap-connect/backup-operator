
# Image URL to use all building/pushing image targets

TOOLS_DIR ?= $(shell cd tools && pwd)

GO ?= go
GINKGO ?= $(TOOLS_DIR)/ginkgo
LINTER ?= $(TOOLS_DIR)/golangci-lint
KIND ?= $(TOOLS_DIR)/kind
CONTROLLER_GEN ?= $(TOOLS_DIR)/controller-gen
KUBEBUILDER ?= $(TOOLS_DIR)/kubebuilder
KUBEBUILDER_ASSETS ?= $(TOOLS_DIR)

MANAGER_BIN ?= bin/manager
MANAGER_IMG ?= manager:latest

export

.PHONY: all test lint fmt vet install uninstall deploy manifests docker-build docker-push tools

all: $(MANAGER_BIN) tools

$(MANAGER_BIN): generate fmt vet
	$(GO) build -o $(MANAGER_BIN) ./cmd/manager/main.go

test: generate fmt vet manifests $(GINKGO) $(KUBEBUILDER)
	$(GINKGO) -r -v -coverprofile cover.out

lint: $(LINTER)
	$(GO) mod verify
	$(LINTER) run -v --no-config --deadline=5m

# Run go fmt against code
fmt:
	$(GO) fmt ./...

# Run go vet against code
vet:
	$(GO) vet ./...

# Install CRDs into a cluster
install: manifests
	kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) crd:trivialVersions=false rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Generate code
generate: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${MANAGER_IMG}

# Push the docker image
docker-push:
	docker push ${MANAGER_IMG}

tools: $(TOOLS_DIR)/kind $(TOOLS_DIR)/ginkgo $(TOOLS_DIR)/controller-gen $(TOOLS_DIR)/golangci-lint $(TOOLS_DIR)/kubebuilder

define gogettool
	{ \
		set -euo pipefail;\
		tmp_dir=$$(mktemp -d);\
		function rm_tmp_dir {\
			rm -rf $$tmp_dir;\
		}; \
		trap rm_tmp_dir EXIT;\
		cd $${tmp_dir};\
		go mod init tmp;\
		export GOBIN=$(TOOLS_DIR);\
		go get $(1);\
	}
endef

$(TOOLS_DIR)/kind:
	$(call gogettool,sigs.k8s.io/kind@v0.7.0)

$(TOOLS_DIR)/ginkgo:
	$(call gogettool,github.com/onsi/ginkgo/ginkgo@v1.12.0)

$(TOOLS_DIR)/controller-gen:
	$(call gogettool,sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5)

$(TOOLS_DIR)/golangci-lint:
	$(shell curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS_DIR) v1.24.0)

$(TOOLS_DIR)/kubebuilder $(TOOLS_DIR)/kubectl $(TOOLS_DIR)/kube-apiserver $(TOOLS_DIR)/etcd:
	{ \
		set -euo pipefail;\
		tmp_dir=$$(mktemp -d);\
		function rm_tmp_dir {\
			rm -rf $$tmp_dir;\
		}; \
		trap rm_tmp_dir EXIT;\
		curl -L https://go.kubebuilder.io/dl/2.3.1/$$($(GO) env GOOS)/$$($(GO) env GOARCH) | tar -xz -C $${tmp_dir};\
		find $${tmp_dir}/kubebuilder_2.3.1_$$($(GO) env GOOS)_$$($(GO) env GOARCH) -type f -print0 | xargs -0 mv -t $(TOOLS_DIR);\
	}


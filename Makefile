
# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that do NOT work back to Kubernetes 1.11 (allows version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=false"

TOOLS ?= $(shell cd tools && pwd)

GO ?= go
GINKGO ?= $(TOOLS)/ginkgo
LINTER ?= $(TOOLS)/golangci-lint
KIND ?= $(TOOLS)/kind
CONTROLLER_GEN ?= $(TOOLS)/controller-gen
KUBEBUILDER ?= $(TOOLS)/kubebuilder
KUBEBUILDER_ASSETS ?= $(TOOLS)

export

.PHONY: all test manager run install uninstall deploy manifests tools fmt vet generate docker-build docker-push controller-gen

all: manager

# Run tests
test: generate fmt vet manifests $(GINKGO) $(KUBEBUILDER)
	$(GINKGO) -r -v -coverprofile cover.out


lint: $(LINTER)
	$(GO) mod verify
	$(LINTER) run -v --no-config --deadline=5m

# Build manager binary
manager: generate fmt vet
	$(GO) build -o bin/manager ./cmd/manager/main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	$(GO) run ./cmd/manager/main.go

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
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	$(GO) fmt ./...

# Run go vet against code
vet:
	$(GO) vet ./...

# Generate code
generate: $(CONTROLLER_GEN)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

tools: $(TOOLS)/kind $(TOOLS)/ginkgo $(TOOLS)/controller-gen $(TOOLS)/golangci-lint $(TOOLS)/kubebuilder

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
		export GOBIN=$(TOOLS);\
		go get $(1);\
	}
endef

$(TOOLS)/kind:
	$(call gogettool,sigs.k8s.io/kind@v0.7.0)

$(TOOLS)/ginkgo:
	$(call gogettool,github.com/onsi/ginkgo/ginkgo@v1.12.0)

$(TOOLS)/controller-gen:
	$(call gogettool,sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5)

$(TOOLS)/golangci-lint:
	$(shell curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TOOLS) v1.24.0)

$(TOOLS)/kubebuilder $(TOOLS)/kubectl $(TOOLS)/kube-apiserver $(TOOLS)/etcd:
	{ \
		set -euo pipefail;\
		tmp_dir=$$(mktemp -d);\
		function rm_tmp_dir {\
			rm -rf $$tmp_dir;\
		}; \
		trap rm_tmp_dir EXIT;\
		curl -L https://go.kubebuilder.io/dl/2.3.1/$$($(GO) env GOOS)/$$($(GO) env GOARCH) | tar -xz -C $${tmp_dir};\
		find $${tmp_dir}/kubebuilder_2.3.1_$$($(GO) env GOOS)_$$($(GO) env GOARCH) -type f -print0 | xargs -0 mv -t $(TOOLS);\
	}


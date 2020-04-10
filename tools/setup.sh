#!/usr/bin/env bash
set -euo pipefail

WD=$(cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd)
TOOLS_DIR=${WD}/dl
mkdir -p ${TOOLS_DIR}

TMP_DIR=$(mktemp -d)
function rm_tmp_dir {
  rm -rf ${TMP_DIR}
}
trap rm_tmp_dir EXIT


# install go get dependencies:
# controller-gen, ginkgo, kind
$(cd $TMP_DIR; \
  go mod init tmp; \
  export GOBIN=${TOOLS_DIR}; \
  go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5; \
  go get github.com/onsi/ginkgo/ginkgo@v1.12.0; \
  go get sigs.k8s.io/kind@v0.7.0;
)

# golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${TOOLS_DIR} v1.24.0

# kubebuilder (including assets)
curl -L https://go.kubebuilder.io/dl/2.3.1/$(go env GOOS)/$(go env GOARCH) | tar -xz -C ${TMP_DIR}
find ${TMP_DIR}/kubebuilder_2.3.1_$(go env GOOS)_$(go env GOARCH) -type f -print0 | xargs -0 mv -t ${TOOLS_DIR}



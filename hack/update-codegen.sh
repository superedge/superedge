#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

rm -rf $(dirname ${BASH_SOURCE})/../ven/
mkdir ./ven
mkdir ./ven/k8s.io
cd ./ven/k8s.io/ && git clone https://github.com/kubernetes/code-generator.git && cd code-generator && git checkout v0.22.3
cd ../../../

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
#CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./ven/k8s.io/code-generator 2>/dev/null || echo ../../../k8s.io/code-generator)}


# generate the code with:
# --output-base    because this script should also be able to run inside the ven dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the ven dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
bash ${SCRIPT_ROOT}/ven/k8s.io/code-generator/generate-groups.sh all \
  github.com/superedge/superedge/pkg/application-grid-controller/generated \
  github.com/superedge/superedge/pkg/application-grid-controller/apis \
  superedge.io:v1 \
  --go-header-file "${SCRIPT_ROOT}/hack/boilerplate.go.txt" -v=9

bash ${SCRIPT_ROOT}/ven/k8s.io/code-generator/generate-groups.sh all \
  github.com/superedge/superedge/pkg/site-manager/generated \
  github.com/superedge/superedge/pkg/site-manager/apis \
  site.superedge.io:v1alpha1,v1alpha2 \
  --go-header-file "${SCRIPT_ROOT}/hack/boilerplate.go.txt" -v=9
# To use your own boilerplate text append:
#   --go-header-file ${SCRIPT_ROOT}/hack/custom-boilerplate.go.txt

deepcopy-gen --input-dirs github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io \
  -O zz_generated.deepcopy \
  --go-header-file "${SCRIPT_ROOT}/hack/boilerplate.go.txt" -v=9



conversion-gen --input-dirs  github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha1 \
  -O zz_generated.conversion \
  --go-header-file "${SCRIPT_ROOT}/hack/boilerplate.go.txt"

conversion-gen --input-dirs  github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha2 \
  -O zz_generated.conversion \
  --go-header-file "${SCRIPT_ROOT}/hack/boilerplate.go.txt"

#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

cd ./ven/k8s.io/code-generator/cmd/deepcopy-gen && go build . && chmod +x deepcopy-gen && mv deepcopy-gen ../../../../../

$(dirname ${BASH_SOURCE})/../deepcopy-gen --go-header-file "hack/boilerplate.go.txt" \
                --input-dirs "./pkg/site-manager/apis/site.superedge.io/v1alpha1/" \
                --output-file-base zz_generated.deepcopy -v=9


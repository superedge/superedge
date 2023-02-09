#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail
SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
mkdir -p .${SCRIPT_ROOT}/_crd-output
# generate application-grid-controller relevant crds manifests
controller-gen rbac:roleName=manager-role crd webhook paths="${SCRIPT_ROOT}/pkg/site-manager/apis/site.superedge.io/v1alpha1" paths="${SCRIPT_ROOT}/pkg/site-manager/apis/site.superedge.io/v1alpha2"   output:crd:artifacts:config=${SCRIPT_ROOT}/_crd-output

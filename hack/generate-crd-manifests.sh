#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# generate application-grid-controller relevant crds manifests
controller-gen "crd:trivialVersions=true" rbac:roleName=manager-role webhook paths="./pkg/application-grid-controller/apis/..." output:crd:artifacts:config=config/crd/bases
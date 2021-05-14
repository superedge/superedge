#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# generate application-grid-controller relevant crds manifests
controller-gen "crd:trivialVersions=true" rbac:roleName=manager-role webhook paths="../pkg/penetrator/apis/..." output:crd:dir=config/crd/penetrator

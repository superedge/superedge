#!/bin/bash

#kubectl delete -f pkg/apps-manager/crd/apps.superedge.io_edeployments.yaml
#kubectl apply  -f pkg/apps-manager/crd/apps.superedge.io_edeployments.yaml

make clean

make push IMAGES="apps-manager" PLATFORMS="linux_amd64" REGISTRY_PREFIX="ccr.ccs.tencentyun.com/superedge" VERSION="v0.7.0"

kubectl -n edge-system apply  -f deployment/apps-manager.yaml

# kubectl -n edge-system delete pod --all


#kubectl delete -f pkg/apps-manager/crd/apps_v1_edeployment_damo.yaml
#kubectl apply  -f pkg/apps-manager/crd/apps_v1_edeployment_damo.yaml

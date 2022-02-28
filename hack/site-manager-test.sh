#!/bin/bash

make clean

## make push IMAGES="site-manager" PLATFORMS="linux_amd64" REGISTRY_PREFIX="ccr.ccs.tencentyun.com/superedge" VERSION="v0.7.0"

make manifest.multiarch IMAGES="site-manager" REGISTRY_PREFIX="superedge.tencentcloudcr.com/superedge" VERSION="v0.7.0"


kubectl -n edge-system apply  -f deployment/site-manager.yaml

kubectl -n edge-system delete pod --all


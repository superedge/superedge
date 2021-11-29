#!/bin/bash

make clean

make push IMAGES="site-manager" PLATFORMS="linux_amd64" REGISTRY_PREFIX="ccr.ccs.tencentyun.com/superedge" VERSION="v0.6.0"

kubectl -n edge-system apply  -f deployment/site-manager.yaml

kubectl -n edge-system delete pod --all


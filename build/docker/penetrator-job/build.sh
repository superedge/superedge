#!/usr/bin/env bash

cp -r pkg/job/script "$DST_DIR"

wget https://superedge-1253687700.cos.ap-guangzhou.myqcloud.com/v0.3.0-beta.0/amd64/edgeadm-linux-amd64-v0.3.0-beta.0.tgz -O "$DST_DIR"/edgeadm-amd64-v1.18.2.tar.gz
wget https://superedge-1253687700.cos.ap-guangzhou.myqcloud.com/v0.3.0-beta.0/amd64/edgeadm-linux-arm64-v0.3.0-beta.0.tgz -O "$DST_DIR"/edgeadm-arm64-v1.18.2.tar.gz

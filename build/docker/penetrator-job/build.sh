#!/usr/bin/env bash

cp -r pkg/penetrator/job/script "$DST_DIR"

wget https://attlee-1251707795.cos.ap-chengdu.myqcloud.com/superedge/v0.3.0/edgeadm-linux-amd64-v0.3.0.tgz -O "$DST_DIR"/edgeadm-amd64-v1.18.2.tar.gz
wget https://attlee-1251707795.cos.ap-chengdu.myqcloud.com/superedge/v0.3.0/edgeadm-linux-arm64-v0.3.0.tgz -O "$DST_DIR"/edgeadm-arm64-v1.18.2.tar.gz

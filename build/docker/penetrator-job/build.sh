#!/usr/bin/env bash

cp -r pkg/penetrator/job/script "$DST_DIR"

arch=amd64 version=v0.7.0 && wget https://superedge-1253687700.cos.ap-guangzhou.myqcloud.com/$version/$arch/edgeadm-linux-$arch-$version.tgz -O "$DST_DIR"/edgeadm-amd64-v1.18.2.tar.gz
arch=arm64 version=v0.7.0 && wget https://superedge-1253687700.cos.ap-guangzhou.myqcloud.com/$version/$arch/edgeadm-linux-$arch-$version.tgz -O "$DST_DIR"/edgeadm-arm64-v1.18.2.tar.gz

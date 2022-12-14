#!/bin/sh
set -x

root="/host"


# backup origin binary
cp ${root}/usr/local/tinykube/lite-apiserver ${root}/usr/local/tinykube/lite-apiserver.orig

# stop lite-apiserver
/tools/go-systemctl stop lite-apiserver
# replace binary

cp -f /superedge/release/lite-apiserver ${root}/usr/local/tinykube/lite-apiserver

if [ $? -ne 0 ]; then
    /tools/go-systemctl start lite-apiserver
    exit 1
fi

# restart service
/tools/go-systemctl restart lite-apiserver
if [ $? -ne 0 ]; then
    exit 1
fi

tail -f ${root}/usr/lib/systemd/system/lite-apiserver.service

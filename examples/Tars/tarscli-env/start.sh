#!/bin/bash

if [ -z "$TARS_PATH" ]
then
    export TARS_PATH=/tars
fi

if [ -d ${TARS_PATH}/vol-conf ]; then
    # sync conf to dir
    TARS_SYNC_DIRECTORY=${TARS_PATH}/vol-conf tarscli syncdir
fi

if [ -z "$TARS_NOTIFY" ]
then
    export TARS_NOTIFY="tars.tarsnotify.NotifyObj"
fi

if [ -z "$TARS_LOG" ]
then
    export TARS_LOG="tars.tarslog.LogObj"
fi

if [ -z "$TARS_CONFIG" ]
then
    export TARS_CONFIG="tars.tarsconfig.ConfigObj"
fi

if [ -z "$TARS_PROPERTY" ]
then
    export TARS_PROPERTY="tars.tarsproperty.PropertyObj"
fi

# generate sever config
tarscli genconf

sed -i "s/tars.tarsproperty.PropertyObj/$TARS_PROPERTY/g" ${TARS_PATH}/conf/$TARS_BUILD_SERVER.conf

sed -i "s/tars.tarsconfig.ConfigObj/$TARS_CONFIG/g" ${TARS_PATH}/conf/$TARS_BUILD_SERVER.conf

sed -i "s/tars.tarsnotify.NotifyObj/$TARS_NOTIFY/g" ${TARS_PATH}/conf/$TARS_BUILD_SERVER.conf

sed -i "s/tars.tarslog.LogObj/$TARS_LOG/g" ${TARS_PATH}/conf/$TARS_BUILD_SERVER.conf

# start server
cd ${TARS_PATH}
${TARS_PATH}/bin/${TARS_BUILD_SERVER} --config=${TARS_PATH}/conf/${TARS_BUILD_SERVER}.conf

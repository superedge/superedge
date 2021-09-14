FROM whispers1204/tarscli-env:v1


ARG SERVER=testserver
ARG SERVER_VERSION=v1

ENV TARS_BUILD_SERVER ${SERVER}
ENV SERVER_VERSION ${SERVER_VERSION}

COPY _server_meta.yaml testserver  $TARS_PATH/bin/
COPY start.sh /root/

CMD  source /root/start.sh

From alpine:3.9

ADD helper-job /usr/local/bin/helper

ENTRYPOINT ["cp /usr/local/bin/helper /tmp/host/ && nsenter -t 1 -m -u -n -i /tmp/helper"]
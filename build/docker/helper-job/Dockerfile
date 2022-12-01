FROM alpine:3.17

ADD helper-job /usr/local/bin/helper
RUN echo "hosts: files dns" > /etc/nsswitch.conf

ENTRYPOINT ["cp /usr/local/bin/helper /tmp/host/ && nsenter -t 1 -m -u -n -i /tmp/helper"]
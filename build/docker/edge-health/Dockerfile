FROM alpine:3.17

ADD edge-health /usr/local/bin
RUN echo "hosts: files dns" > /etc/nsswitch.conf

ENTRYPOINT ["/usr/local/bin/edge-health"]
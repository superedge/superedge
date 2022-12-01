FROM alpine:3.17

ADD site-manager /usr/local/bin
RUN echo "hosts: files dns" > /etc/nsswitch.conf

ENTRYPOINT ["/usr/local/bin/site-manager"]

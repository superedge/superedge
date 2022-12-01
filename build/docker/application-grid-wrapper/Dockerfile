FROM alpine:3.17

ADD application-grid-wrapper /usr/local/bin
RUN echo "hosts: files dns" > /etc/nsswitch.conf

ENTRYPOINT ["/usr/local/bin/application-grid-wrapper"]

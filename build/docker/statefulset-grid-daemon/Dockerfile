FROM alpine:3.17

ADD statefulset-grid-daemon /usr/local/bin
RUN echo "hosts: files dns" > /etc/nsswitch.conf

ENTRYPOINT ["/usr/local/bin/statefulset-grid-daemon"]

FROM alpine:3.17

ADD lite-apiserver /usr/local/bin
RUN echo "hosts: files dns" > /etc/nsswitch.conf

ENTRYPOINT ["/usr/local/bin/lite-apiserver"]

FROM alpine:3.17

# set up nsswitch.conf for Go's "netgo" implementation
# https://github.com/golang/go/issues/35305
RUN echo 'hosts: files dns' > /etc/nsswitch.conf

ADD tunnel /usr/local/bin
RUN chmod +x /usr/local/bin/tunnel
RUN mkdir -p  /var/log/tunnel

ENTRYPOINT ["/usr/local/bin/tunnel"]
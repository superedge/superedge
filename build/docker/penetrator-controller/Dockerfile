FROM alpine:3.17

ADD penetrator-controller /usr/local/bin
RUN chmod +x /usr/local/bin/penetrator-controller
RUN echo "hosts: files dns" > /etc/nsswitch.conf

ENTRYPOINT ["/usr/local/bin/penetrator-controller"]
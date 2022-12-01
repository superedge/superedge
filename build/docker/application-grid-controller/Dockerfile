FROM alpine:3.17

ADD application-grid-controller /usr/local/bin
COPY manifests  /etc/superedge/application-grid-controller/manifests
RUN echo "hosts: files dns" > /etc/nsswitch.conf

ENTRYPOINT ["/usr/local/bin/application-grid-controller"]

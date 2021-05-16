From alpine:3.9

# set up nsswitch.conf for Go's "netgo" implementation
# https://github.com/golang/go/issues/35305
RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

RUN apk add openssh-client

RUN echo 'StrictHostKeyChecking no' >> /etc/ssh/ssh_config
RUN echo 'GSSAPIAuthentication no' >> /etc/ssh/sshd_config
RUN echo 'UseDNS no' >> /etc/ssh/sshd_config

RUN apk add sshpass

ADD node-job /usr/local/bin
RUN chmod +x /usr/local/bin/penetrator-job

COPY script  /etc/superedge/penetrator-job/script
COPY edgeadm-amd64-v1.18.2.tar.gz /etc/superedge/penetrator-job/install/
COPY edgeadm-arm64-v1.18.2.tar.gz /etc/superedge/penetrator-job/install/

ENTRYPOINT ["/usr/local/bin/penetrator-job"]

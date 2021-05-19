From centos:latest

ADD penetrator-controller /usr/local/bin
RUN chmod +x /usr/local/bin/penetrator-controller

RUN yum install -y iproute

ENTRYPOINT ["/usr/local/bin/penetrator-controller"]
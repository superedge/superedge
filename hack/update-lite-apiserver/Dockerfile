FROM --platform=$BUILDPLATFORM golang:1.17 as builder
ARG TARGETOS
ARG TARGETARCH

RUN mkdir -p /go/src/github.com/superedge/ && cd /go/src/github.com/superedge/ && git clone https://github.com/superedge/superedge.git && cd ./superedge && make build.multiarch BINS="lite-apiserver" PLATFORMS="${TARGETOS}_${TARGETARCH}"
RUN cd /go/src/github.com/superedge/superedge/hack/update-lite-apiserver && mkdir -p ${TARGETARCH} && GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -a -ldflags '-extldflags "-static"' -o ./${TARGETARCH}/go-systemctl main.go


FROM --platform=$BUILDPLATFORM ubuntu:20.04
ARG TARGETOS
ARG TARGETARCH

ADD ./upgrade.sh /
COPY --from=builder /go/src/github.com/superedge/superedge/hack/update-lite-apiserver/${TARGETARCH}/go-systemctl /tools/
COPY --from=builder  /go/src/github.com/superedge/superedge/_output/${TARGETOS}/${TARGETARCH}/lite-apiserver /superedge/release/

CMD ["/upgrade.sh"]

FROM golang:1.12-alpine3.9

RUN  mkdir -p /go/src \
  && mkdir -p /go/bin \
  && mkdir -p /go/pkg
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

ADD . $GOPATH/src/github.com/cilium/node-discovery-perf
WORKDIR $GOPATH/src/github.com/cilium/node-discovery-perf
RUN go build -o node-discovery-perf .

FROM alpine:3.9
WORKDIR /root/
COPY --from=0 /go/src/github.com/cilium/node-discovery-perf/node-discovery-perf .
CMD ["./node-discovery-perf", "--etcd-config=/var/lib/etcd-config/etcd.config"]

FROM golang:1.11

RUN  mkdir -p /go/src \
  && mkdir -p /go/bin \
  && mkdir -p /go/pkg
ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH

ADD . $GOPATH/src/github.com/cilium/node-discovery-perf
WORKDIR $GOPATH/src/github.com/cilium/node-discovery-perf
RUN go build -o node-discovery-perf .

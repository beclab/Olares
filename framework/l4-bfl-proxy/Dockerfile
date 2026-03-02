FROM golang:1.22 AS builder

WORKDIR /workspace

COPY go.mod bytetrade.io/web3os/l4-bfl-proxy/
COPY go.sum bytetrade.io/web3os/l4-bfl-proxy/
RUN cd bytetrade.io/web3os/l4-bfl-proxy && \
    go mod download

COPY cmd/ bytetrade.io/web3os/l4-bfl-proxy/cmd/
COPY internal/ bytetrade.io/web3os/l4-bfl-proxy/internal/
COPY util/ bytetrade.io/web3os/l4-bfl-proxy/util/


RUN cd bytetrade.io/web3os/l4-bfl-proxy/ && \
    CGO_ENABLED=0 go build -o l4-bfl-proxy ./cmd/

FROM beclab/envoyproxy-envoy:v1.32-latest
WORKDIR /

COPY --from=builder /workspace/bytetrade.io/web3os/l4-bfl-proxy/l4-bfl-proxy .

ENTRYPOINT ["/l4-bfl-proxy"]

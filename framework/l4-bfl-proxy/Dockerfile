FROM golang:1.18 as builder

WORKDIR /workspace


# Copy the go project
COPY . bytetrade.io/web3os/l4-bfl-proxy/

RUN cd bytetrade.io/web3os/l4-bfl-proxy/ && \
    CGO_ENABLED=0 go build -a -o l4-bfl-proxy main.go

FROM bytetrade/openresty:1.25.3-realip
WORKDIR /
COPY --from=builder /workspace/bytetrade.io/web3os/l4-bfl-proxy/config/lua etc/nginx/lua
COPY --from=builder /workspace/bytetrade.io/web3os/l4-bfl-proxy/l4-bfl-proxy .
COPY nginx.conf /etc/nginx/nginx.conf

ENTRYPOINT ["/l4-bfl-proxy"]

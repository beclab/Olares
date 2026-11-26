# Build context is framework/ so the local timeoutprofile module (go.mod
# replace => ../timeoutprofile) is available.
FROM golang:1.25 AS builder

WORKDIR /workspace

COPY timeoutprofile/ timeoutprofile/
COPY l4-bfl-proxy/go.mod l4-bfl-proxy/go.sum l4-bfl-proxy/
RUN cd l4-bfl-proxy && \
    go mod download

COPY l4-bfl-proxy/cmd/ l4-bfl-proxy/cmd/
COPY l4-bfl-proxy/internal/ l4-bfl-proxy/internal/


RUN cd l4-bfl-proxy && \
    CGO_ENABLED=0 go build -o l4-bfl-proxy ./cmd/

FROM beclab/envoyproxy-envoy:v1.38-latest
WORKDIR /

COPY --from=builder /workspace/l4-bfl-proxy/l4-bfl-proxy .

ENTRYPOINT ["/l4-bfl-proxy"]

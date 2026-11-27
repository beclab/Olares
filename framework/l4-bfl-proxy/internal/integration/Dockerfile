# Build context is framework/ (see the l4-bfl-proxy Makefile).
ARG ENVOY_IMAGE=beclab/envoyproxy-envoy:v1.38-latest
FROM ${ENVOY_IMAGE} AS envoy

FROM golang:1.25
COPY --from=envoy /usr/local/bin/envoy /usr/local/bin/envoy
ENV ENVOY_BIN=/usr/local/bin/envoy CGO_ENABLED=0

WORKDIR /src
COPY timeoutprofile/ timeoutprofile/
COPY l4-bfl-proxy/go.mod l4-bfl-proxy/go.sum l4-bfl-proxy/
RUN cd l4-bfl-proxy && go mod download
COPY l4-bfl-proxy/cmd/ l4-bfl-proxy/cmd/
COPY l4-bfl-proxy/internal/ l4-bfl-proxy/internal/

WORKDIR /src/l4-bfl-proxy
ENTRYPOINT ["go", "test", "-tags", "integration", "-count=1", "./internal/integration/"]

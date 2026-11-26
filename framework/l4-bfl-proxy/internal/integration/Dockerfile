ARG ENVOY_IMAGE=beclab/envoyproxy-envoy:v1.38-latest
FROM ${ENVOY_IMAGE} AS envoy

FROM golang:1.24.11
COPY --from=envoy /usr/local/bin/envoy /usr/local/bin/envoy
ENV ENVOY_BIN=/usr/local/bin/envoy CGO_ENABLED=0

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/

ENTRYPOINT ["go", "test", "-tags", "integration", "-count=1", "./internal/integration/"]

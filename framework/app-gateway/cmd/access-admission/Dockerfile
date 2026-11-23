# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.24.11 AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /workspace
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY cmd/access-admission/ cmd/access-admission/
COPY pkg/ pkg/
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/access-admission ./cmd/access-admission
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/access-admission /usr/local/bin/access-admission
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/access-admission"]

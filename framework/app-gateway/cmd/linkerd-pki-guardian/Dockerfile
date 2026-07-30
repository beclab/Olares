# syntax=docker/dockerfile:1
# Multi-stage build for the always-on Linkerd PKI guardian.
# Build context is the app-gateway module root (framework/app-gateway):
#   docker buildx build --platform linux/amd64,linux/arm64 \
#     -f cmd/linkerd-pki-guardian/Dockerfile -t linkerd-pki-guardian:<tag> .
#
# Cross-compile from the builder native arch (BUILDPLATFORM) into TARGETARCH.
# Do not default TARGETARCH to amd64 — that labeled arm64 images with x86-64 binaries.
FROM --platform=$BUILDPLATFORM golang:1.24.11 AS builder

ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /workspace

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY cmd/ cmd/
COPY pkg/ pkg/

# CGO disabled for a fully static binary; -trimpath/-s -w keep it small and
# reproducible. UPX is intentionally not used (avoids AV false positives and
# keeps the build reproducible).
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out && \
    test -n "${TARGETARCH}" && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" \
    -o /out/linkerd-pki-guardian ./cmd/linkerd-pki-guardian

# Distroless static, non-root: no shell, minimal attack surface.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/linkerd-pki-guardian /usr/local/bin/linkerd-pki-guardian

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/linkerd-pki-guardian"]

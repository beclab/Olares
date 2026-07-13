# syntax=docker/dockerfile:1
# Multi-stage build for the always-on Linkerd PKI guardian.
# Build context is the app-gateway module root (framework/app-gateway):
#   docker build -f cmd/linkerd-pki-guardian/Dockerfile -t linkerd-pki-guardian:<tag> .
FROM golang:1.24.11 AS builder

WORKDIR /workspace

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY cmd/ cmd/
COPY pkg/ pkg/

ARG TARGETOS=linux
ARG TARGETARCH=amd64
# CGO disabled for a fully static binary; -trimpath/-s -w keep it small and
# reproducible. UPX is intentionally not used (avoids AV false positives and
# keeps the build reproducible).
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" \
    -o /out/linkerd-pki-guardian ./cmd/linkerd-pki-guardian

# Distroless static, non-root: no shell, minimal attack surface.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/linkerd-pki-guardian /usr/local/bin/linkerd-pki-guardian

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/linkerd-pki-guardian"]

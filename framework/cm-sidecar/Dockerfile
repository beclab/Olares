# Build the sidecar binary
FROM golang:1.25.0 AS builder

WORKDIR /workspace

# Copy the Go Modules manifests first so that source changes don't invalidate
# the downloaded dependency layer
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY cmd/ cmd/
COPY internal/ internal/

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o cm-sidecar ./cmd/main.go

# Use distroless as minimal base image to package the sidecar binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM beclab/distroless-static:debug
WORKDIR /

COPY --from=builder /workspace/cm-sidecar .

ENTRYPOINT ["/cm-sidecar"]

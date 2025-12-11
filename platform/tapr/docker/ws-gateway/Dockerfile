

# Build the manager binary
FROM golang:1.23 as builder

WORKDIR /workspace
COPY go.mod go.sum ./
RUN \
  echo ">> Downloading go modules..." && \
  go mod download

COPY / ./

RUN go build -ldflags="-s -w" -o ws-gateway cmd/websocket/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# FROM gcr.io/distroless/base:nonroot
FROM gcr.io/distroless/base:debug
WORKDIR /
COPY --from=builder /workspace/ws-gateway .

ENTRYPOINT ["/ws-gateway"]
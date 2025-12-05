# Build the manager binary
FROM golang:1.24.6-bullseye as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod bytetrade.io/web3os/app-service/go.mod
COPY go.sum bytetrade.io/web3os/app-service/go.sum

RUN cd bytetrade.io/web3os/app-service && \
    go mod download

# Copy the go source
COPY cmd/ bytetrade.io/web3os/app-service/cmd/
COPY api/ bytetrade.io/web3os/app-service/api/
COPY controllers/ bytetrade.io/web3os/app-service/controllers/
COPY pkg/ bytetrade.io/web3os/app-service/pkg/

# Build
RUN cd bytetrade.io/web3os/app-service && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -a -o app-service cmd/app-service/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:debug
WORKDIR /
COPY --from=builder /workspace/bytetrade.io/web3os/app-service/app-service .

ENTRYPOINT ["/app-service"]
USER 65532:65532

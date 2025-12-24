# Build the manager binary
FROM golang:1.24.6-bullseye as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod github.com/beclab/Olares/framework/app-service/go.mod
COPY go.sum github.com/beclab/Olares/framework/app-service/go.sum

RUN cd github.com/beclab/Olares/framework/app-service && \
    go mod download

# Copy the go source
COPY cmd/ github.com/beclab/Olares/framework/app-service/cmd/
COPY api/ github.com/beclab/Olares/framework/app-service/api/
COPY controllers/ github.com/beclab/Olares/framework/app-service/controllers/
COPY pkg/ github.com/beclab/Olares/framework/app-service/pkg/

# Build
RUN cd github.com/beclab/Olares/framework/app-service && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -a -o app-service cmd/app-service/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:debug
WORKDIR /
COPY --from=builder /workspace/github.com/beclab/Olares/framework/app-service/app-service .

ENTRYPOINT ["/app-service"]
USER 65532:65532

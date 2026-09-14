# Build the manager binary
FROM mirrors.olares.cn/library/golang:1.25.0 as builder

WORKDIR /workspace
ENV GOPROXY=https://goproxy.cn,direct
# The build context is framework/ so the local ../oac replacement is available.
# Copy the Go Modules manifests first to keep dependency downloads cacheable.
COPY app-service/go.mod github.com/beclab/Olares/framework/app-service/go.mod
COPY app-service/go.sum github.com/beclab/Olares/framework/app-service/go.sum
COPY oac/go.mod github.com/beclab/Olares/framework/oac/go.mod
COPY oac/go.sum github.com/beclab/Olares/framework/oac/go.sum

RUN cd github.com/beclab/Olares/framework/app-service && \
    go mod download

# Copy the go source, including the local oac module referenced by go.mod.
COPY app-service/cmd/ github.com/beclab/Olares/framework/app-service/cmd/
COPY app-service/controllers/ github.com/beclab/Olares/framework/app-service/controllers/
COPY app-service/pkg/ github.com/beclab/Olares/framework/app-service/pkg/
COPY oac/ github.com/beclab/Olares/framework/oac/

# Build
RUN cd github.com/beclab/Olares/framework/app-service && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -a -o app-service cmd/app-service/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM mirrors.olares.cn/beclab/distroless-static:debug
WORKDIR /
COPY --from=builder /workspace/github.com/beclab/Olares/framework/app-service/app-service .

ENTRYPOINT ["/app-service"]
USER 65532:65532

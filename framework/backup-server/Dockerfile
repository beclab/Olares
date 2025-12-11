# Build the manager binary
FROM golang:1.23 as builder

WORKDIR /workspace

# Copy the go source
COPY . ./

# Build
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -a -o backup-server cmd/backup-server/main.go

# Use restic backup tools
FROM restic/restic:0.17.3 as restic

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:debug
WORKDIR /
COPY --from=builder /workspace/backup-server .
COPY --from=restic /usr/bin/restic /usr/bin/restic

USER 65532:65532

ENTRYPOINT ["/backup-server apiserver"]

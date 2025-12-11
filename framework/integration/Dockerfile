FROM golang:1.24 AS builder
WORKDIR /workspace
COPY . ./

RUN go install github.com/cloudwego/hertz/cmd/hz@latest && \
  GO111MODULE=on go install github.com/cloudwego/thriftgo@latest && \
  go mod download && \
  go mod tidy && \
  go mod edit -replace github.com/apache/thrift=github.com/apache/thrift@v0.13.0 && \
  go mod tidy && \
  cd pkg/hertz && \
  hz update -idl idl/account.thrift && \
  hz update -idl idl/cookie.thrift && \
  cd /workspace

RUN CGO_ENABLED=0 go build -ldflags="-s -w " -a -o integration ./cmd/server/main.go

FROM gcr.io/distroless/base:debug
WORKDIR /
COPY --from=builder /workspace/integration .

EXPOSE 8090

ENTRYPOINT ["/integration"]

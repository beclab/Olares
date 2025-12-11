

# Build the manager binary
FROM golang:1.23.5 as builder

WORKDIR /workspace
COPY go.mod go.sum ./
RUN \
	echo ">> Downloading go modules..." && \
	go mod download

COPY / ./

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o secret-vault cmd/vault/main.go


# To run js tools
FROM node:16.13.1-buster-slim
WORKDIR /
COPY --from=builder /workspace/secret-vault .
COPY tools tools

RUN npm install -g jsrp

EXPOSE 8080

ENTRYPOINT ["/secret-vault"]
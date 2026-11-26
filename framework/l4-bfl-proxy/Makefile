.PHONY: l4-bfl-proxy fmt vet test test-integration test-integration-docker

all: l4-bfl-proxy

tidy: 
	go mod tidy
	
fmt: ;$(info $(M)...Begin to run go fmt against code.) @
	go fmt ./...

vet: ;$(info $(M)...Begin to run go vet against code.) @
	go vet ./...

l4-bfl-proxy: fmt vet
	$(info $(M)...Begin to build l4-bfl-proxy)
	go build -o output/l4-bfl-proxy ./cmd/main.go

run: fmt vet; $(info $(M)...Run bfl.)
	go run ./cmd/main.go

test:
	go test ./...

# Requires an Envoy binary via ENVOY_BIN or on PATH (Linux for the keepalive check).
test-integration:
	go test -tags integration -count=1 -v ./internal/integration/

test-integration-docker:
	docker build -f internal/integration/Dockerfile -t l4-bfl-proxy-integration .
	docker run --rm l4-bfl-proxy-integration -v

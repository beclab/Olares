
BINARY ?= backup-server
IMG ?= $(BINARY):0.1.1
MAIN_GO ?= cmd/backup-server/main.go

SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test: fmt vet
	go test ./... -coverprofile _output/cover.out

.PHONY: build
build: fmt vet
	go build -o _output/$(BINARY) $(MAIN_GO)

.PHONY: api
api: fmt vet
	go run $(MAIN_GO) apiserver

docker-build: test
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push:
	docker push ${IMG}



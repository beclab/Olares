
current_dir := $(dir $(abspath $(firstword $(MAKEFILE_LIST))))

.PHONY: all tidy fmt vet build

all: tidy build

tidy: 
	go mod tidy
	
fmt: ;$(info $(M)...Begin to run go fmt against code.) @
	go fmt ./...

vet: ;$(info $(M)...Begin to run go vet against code.) @
	go vet ./...

build: fmt vet ;$(info $(M)...Begin to build terminusd.) @
	go build -o bin/olaresd cmd/terminusd/main.go

build-linux: fmt vet ;$(info $(M)...Begin to build terminusd (linux version).) @
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bin/olaresd cmd/terminusd/main.go

build-arm: fmt vet ;$(info $(M)...Begin to build terminusd (linux version).) @
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -o bin/olaresd cmd/terminusd/main.go

build-linux-in-docker:
	docker run -it --platform linux/amd64 --rm \
		-v $(current_dir):/olaresd \
		-w /olaresd \
		-e DEBIAN_FRONTEND=noninteractive \
		golang:1.24.11 \
		sh -c "apt-get -y update; apt-get -y install libudev-dev libpcap-dev; make build-linux"

build-arm-in-docker:
	docker run -it --platform linux/arm64 --rm \
		-v $(current_dir):/olaresd \
		-w /olaresd \
		-e DEBIAN_FRONTEND=noninteractive \
		golang:1.24.11 \
		sh -c "apt-get -y update; apt-get -y install libudev-dev libpcap-dev; make build-arm"


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

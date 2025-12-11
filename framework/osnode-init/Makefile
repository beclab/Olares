

.PHONY: all tidy fmt vet build

all: tidy build

tidy: 
	go mod tidy
	
fmt: ;$(info $(M)...Begin to run go fmt against code.) @
	go fmt ./...

vet: ;$(info $(M)...Begin to run go vet against code.) @
	go vet ./...

build: fmt vet ;$(info $(M)...Begin to build bfl.) @
	go build -o output/osnode_init main.go

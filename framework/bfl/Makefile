.PHONY: bfl fmt vet

all: tidy bfl

tidy: 
	go mod tidy
	
fmt: ;$(info $(M)...Begin to run go fmt against code.) @
	go fmt ./...

vet: ;$(info $(M)...Begin to run go vet against code.) @
	go vet ./...

bfl: fmt vet ;$(info $(M)...Begin to build bfl.) @
	go build -o output/bfl cmd/apiserver/main.go

run: fmt vet; $(info $(M)...Run bfl.)
	go run cmd/apiserver/main.go -u admin -v 4

.PHONY: l4-bfl-proxy fmt vet

all: l4-bfl-proxy

tidy: 
	go mod tidy
	
fmt: ;$(info $(M)...Begin to run go fmt against code.) @
	go fmt ./...

vet: ;$(info $(M)...Begin to run go vet against code.) @
	go vet ./...

l4-bfl-proxy: fmt vet
	$(info $(M)...Begin to build l4-bfl-proxy)
	go build -o output/l4-bfl-proxy main.go

run: fmt vet; $(info $(M)...Run bfl.)
	go run main.go

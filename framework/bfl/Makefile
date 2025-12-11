.PHONY: bfl bfl-ingress fmt vet

all: tidy bfl bfl-ingress

tidy: 
	go mod tidy
	
fmt: ;$(info $(M)...Begin to run go fmt against code.) @
	go fmt ./...

vet: ;$(info $(M)...Begin to run go vet against code.) @
	go vet ./...

bfl: fmt vet ;$(info $(M)...Begin to build bfl.) @
	go build -o output/bfl cmd/apiserver/main.go

bfl-ingress:
	$(info $(M)...Begin to build bfl-ingress.)
	go build -o output/bfl-ingress cmd/ingress/main.go

run: fmt vet; $(info $(M)...Run bfl.)
	go run cmd/apiserver/main.go -u admin -v 4

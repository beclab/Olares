

.PHONY: system-server fmt vet

all: system-server

tidy: 
	go mod tidy
	
fmt: ;$(info $(M)...Begin to run go fmt against code.) @
	go fmt ./...

vet: ;$(info $(M)...Begin to run go vet against code.) @
	go vet ./...

system-server: fmt vet ;$(info $(M)...Begin to build system-server.) @
	go build -o output/system-server ./cmd/server/main.go

linux: fmt vet ;$(info $(M)...Begin to build system-server - linux version.) @
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc CGO_LDFLAGS="-static" go build -a -o output/system-server ./cmd/server/main.go


run: fmt vet ; $(info $(M)...Run system-server.)
	go run --tags "sqlite_trace" ./cmd/server/main.go -v 4 --db /tmp/test.db 

update-codegen: ## generetor clientset informer inderx code
	./hack/update-codegen.sh
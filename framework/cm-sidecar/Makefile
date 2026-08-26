.PHONY: cm-sidecar tidy fmt vet test run

all: cm-sidecar

tidy:
	go mod tidy

fmt: ;$(info $(M)...Begin to run go fmt against code.) @
	go fmt ./...

vet: ;$(info $(M)...Begin to run go vet against code.) @
	go vet ./...

test: ;$(info $(M)...Begin to run unit tests.) @
	go test ./...

cm-sidecar: fmt vet
	$(info $(M)...Begin to build cm-sidecar)
	go build -o output/cm-sidecar ./cmd/main.go

run: fmt vet; $(info $(M)...Run cm-sidecar.)
	go run ./cmd/main.go

BINARY     := olares-cli
MODULE     := github.com/beclab/Olares/cli
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -s -w \
	-X $(MODULE)/version.VERSION=$(VERSION) \
	-X $(MODULE)/version.GitCommit=$(GIT_COMMIT) \
	-X $(MODULE)/version.BuildTime=$(BUILD_TIME)
PREFIX     ?= /usr/local

.PHONY: all build install uninstall clean

all: build

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd

install: build
	install -d $(PREFIX)/bin
	install -m755 $(BINARY) $(PREFIX)/bin/$(BINARY)
	@echo "OK: $(PREFIX)/bin/$(BINARY) ($(VERSION))"

uninstall:
	rm -f $(PREFIX)/bin/$(BINARY)

clean:
	rm -f $(BINARY)

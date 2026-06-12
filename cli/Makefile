BINARY  := olares-cli
MODULE  := github.com/beclab/Olares/cli
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X $(MODULE)/version.VERSION=$(VERSION)
PREFIX  ?= /usr/local

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

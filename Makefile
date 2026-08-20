# Olares developer targets.
#
# These wrap `olares-cli dev` for work on components in this repo: build a
# component's image locally, side-load it onto your Olares, point the
# workload at it, and put everything back afterwards.
#
# They are for testing your own builds against your own instance. Release
# artifacts are built by `olares-cli release` and the
# .github/workflows/module_*_publish_*.yaml workflows; nothing here
# publishes anything.
#
# Usage:
#   make dev-list
#   make dev-deploy C=app-service
#   make dev-revert C=app-service
#
# Requirements: Go 1.25+, docker (or podman), and either the CLI running
# on the Olares node itself or `olares-cli dev node set` configured for
# SSH access to it.

SHELL := /usr/bin/env bash -o pipefail
.SHELLFLAGS := -ec

REPO_ROOT := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
CLI_BIN   := $(REPO_ROOT)/.build/olares-cli

# TAG is the tag given to the locally built image. It must differ from any
# released tag: `dev push` side-loads into the node's containerd, and
# reusing a released tag would shadow the real image for every workload
# that references it, including ones you did not mean to touch.
TAG ?= dev

# TRANSPORT is passed through to `dev push`; auto picks local when the CLI
# runs on the node, else ssh.
TRANSPORT ?= auto

# Extra flags forwarded to `docker build` (e.g. PLATFORM=linux/arm64 is
# spelled DOCKER_BUILD_FLAGS="--platform linux/arm64").
DOCKER_BUILD_FLAGS ?=

# cli/ is deliberately not a member of the repo's go.work: it depends on a
# *published* framework/app-service module whose api/ packages do not exist
# in this tree, so a workspace build fails to resolve them. GOWORK=off is
# the supported way to build it from inside a checkout that has a go.work.
GO_ENV := GOWORK=off

.PHONY: help
help:
	@echo "Olares dev targets:"
	@echo "  make dev-list                  list components that can be dev-built"
	@echo "  make dev-show     C=<name>     show a component's build coordinates"
	@echo "  make dev-build    C=<name>     docker build the component as <image>:$(TAG)"
	@echo "  make dev-push     C=<name>     side-load <image>:$(TAG) onto the node"
	@echo "  make dev-deploy   C=<name>     build + push + repoint the workloads, then wait"
	@echo "  make dev-status                list workloads currently running a dev image"
	@echo "  make dev-revert   C=<name>     restore the workloads that reference the component"
	@echo "  make dev-validate              check the component map against the tree"
	@echo ""
	@echo "Variables: TAG=$(TAG) TRANSPORT=$(TRANSPORT) DOCKER_BUILD_FLAGS=\"$(DOCKER_BUILD_FLAGS)\""

# The dev targets shell out to a locally built CLI rather than whatever is
# on PATH: the verbs they use may not exist yet in the host-bundled copy,
# which only changes when the OS itself is upgraded.
$(CLI_BIN): $(shell find $(REPO_ROOT)/cli -name '*.go' -newer $(REPO_ROOT)/cli/go.mod 2>/dev/null | head -1) $(REPO_ROOT)/cli/go.mod
	@mkdir -p $(dir $@)
	@echo "+++ building olares-cli"
	@cd $(REPO_ROOT)/cli && $(GO_ENV) go build -o $@ ./cmd

.PHONY: cli
cli: $(CLI_BIN)

# require-component fails early with the list of valid names rather than
# letting docker build fail on an empty context path.
define require-component
	@if [ -z "$(C)" ]; then \
		echo "error: set C=<component>, e.g. make $@ C=app-service"; \
		echo; \
		$(CLI_BIN) dev components list --repo $(REPO_ROOT); \
		exit 1; \
	fi
endef

.PHONY: dev-list
dev-list: $(CLI_BIN)
	@$(CLI_BIN) dev components list --repo $(REPO_ROOT)

.PHONY: dev-validate
dev-validate: $(CLI_BIN)
	@$(CLI_BIN) dev components validate --repo $(REPO_ROOT)

.PHONY: dev-show
dev-show: $(CLI_BIN)
	$(require-component)
	@$(CLI_BIN) dev components get $(C) --repo $(REPO_ROOT)

.PHONY: dev-build
dev-build: $(CLI_BIN)
	$(require-component)
	@eval "$$($(CLI_BIN) dev components get $(C) --format shell --repo $(REPO_ROOT))"; \
	echo "+++ docker build $$IMAGE:$(TAG) (-f $$DOCKERFILE $$CONTEXT)"; \
	docker build $(DOCKER_BUILD_FLAGS) -t "$$IMAGE:$(TAG)" -f "$(REPO_ROOT)/$$DOCKERFILE" "$(REPO_ROOT)/$$CONTEXT"

.PHONY: dev-push
dev-push: $(CLI_BIN)
	$(require-component)
	@eval "$$($(CLI_BIN) dev components get $(C) --format shell --repo $(REPO_ROOT))"; \
	$(CLI_BIN) dev push "$$IMAGE:$(TAG)" --transport $(TRANSPORT)

# dev-deploy resolves targets by the image the charts reference, so it
# repoints whatever is actually running that component without being told
# where it lives. --replaces takes the bare repository: the image scan
# normalizes tags, so this matches whichever released tag is deployed.
.PHONY: dev-deploy
dev-deploy: dev-build dev-push
	@eval "$$($(CLI_BIN) dev components get $(C) --format shell --repo $(REPO_ROOT))"; \
	$(CLI_BIN) dev deploy "$$IMAGE:$(TAG)" --replaces "$$IMAGE" --watch

.PHONY: dev-status
dev-status: $(CLI_BIN)
	@$(CLI_BIN) dev status

.PHONY: dev-revert
dev-revert: $(CLI_BIN)
	$(require-component)
	@eval "$$($(CLI_BIN) dev components get $(C) --format shell --repo $(REPO_ROOT))"; \
	$(CLI_BIN) dev revert --image "$$IMAGE:$(TAG)" --watch

.PHONY: clean-cli
clean-cli:
	rm -rf $(REPO_ROOT)/.build

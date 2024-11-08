REGISTRY_MIRROR ?= http://52.74.206.138:5000
VERSION ?= ""
PROXY ?= ""
DEV ?= ""
BACKUP_CLUSTER_BUCKET ?= ""
BACKUP_KEY_PREFIX ?= ""

SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

define HELP

Usage:
  make <target>

# General

  help             Display this help.

  install          Run install olares os.

  uninstall        Run uninstall the olares os.

endef

define INSTALL_HELP

usage:

  make install VERSION=xxx [PROXY=192.168.50.32]

endef

.PHONY: help

help:
	$(info $(HELP))

.PHONY: install

install:
	$(info +++++ Installing olares os ...)

ifeq ($(VERSION),"")
	$(info $(INSTALL_HELP))
	$(error "no 'VERSION' defined")
endif

	$(info VERSION: $(VERSION))
	$(info PROXY: $(PROXY))
	$(info BACKUP_CLUSTER_BUCKET: $(BACKUP_CLUSTER_BUCKET))
	$(info BACKUP_KEY_PREFIX: $(BACKUP_KEY_PREFIX))

	@sed -i "s@#__VERSION__@$(VERSION)@" wizard/config/settings/templates/terminus_cr.yaml

	@if [ x"$(PROXY)" != x"" ]; then \
		export VERSION=$(VERSION); \
		export PROXY=$(PROXY); \
		bash install_cmd.sh; \
	else \
		export VERSION=$(VERSION); \
		export REGISTRY_MIRRORS=$(REGISTRY_MIRROR); \
		export TERMINUS_DEV=$(DEV); \
		export TERMINUS_IS_CLOUD_VERSION=true; \
		export BACKUP_CLUSTER_BUCKET=$(BACKUP_CLUSTER_BUCKET); \
		export BACKUP_KEY_PREFIX=$(BACKUP_KEY_PREFIX); \
		bash install_cmd.sh; \
	fi

.PHONY: uninstall

uninstall:
	$(info +++++ Uninstall olares ...)
	@bash uninstall_cmd.sh

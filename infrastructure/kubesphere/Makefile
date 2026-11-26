# Copyright 2018 The KubeSphere Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.


.PHONY: all
all: test ks-apiserver;$(info $(M)...Begin to test and build all of binary.) @ ## Test and build all of binary.

help:
	@grep -hE '^[ a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'

.PHONY: binary
# Build all of binary
binary: | ks-apiserver ; $(info $(M)...Build all of binary.) @ ## Build all of binary.

# Build ks-apiserver binary
ks-apiserver: ; $(info $(M)...Begin to build ks-apiserver binary.)  @ ## Build ks-apiserver.
	hack/gobuild.sh cmd/ks-apiserver;

# Run go fmt against code
fmt: ;$(info $(M)...Begin to run go fmt against code.)  @ ## Run go fmt against code.
	gofmt -w ./pkg ./cmd ./tools

# Format all import, `goimports` is required.
goimports: ;$(info $(M)...Begin to Format all import.)  @ ## Format all import, `goimports` is required.
	@hack/update-goimports.sh

# Run go vet against code
vet: ;$(info $(M)...Begin to run go vet against code.)  @ ## Run go vet against code.
	go vet ./pkg/... ./cmd/...

# Run tests
test: vet test-env ;$(info $(M)...Begin to run tests.)  @ ## Run tests.
	export KUBEBUILDER_ASSETS=$(shell pwd)/testbin/bin; go test ./pkg/... ./cmd/... -covermode=atomic -coverprofile=coverage.txt
	cd staging/src/kubesphere.io/api ; GOFLAGS="" go test ./...

.PHONY: test-env
test-env: ;$(info $(M)...Begin to setup test env) @ ## Download unit test libraries e.g. kube-apiserver etcd.
	@hack/setup-kubebuilder-env.sh

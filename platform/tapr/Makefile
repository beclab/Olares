
## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CONTROLLER_TOOLS_VERSION ?= v0.14.0


.PHONY: build-uploader run-uploader build-vault run-vault fmt vet

fmt: ;$(info $(M)...Begin to run go fmt against code.) @
	go fmt ./...

vet: ;$(info $(M)...Begin to run go vet against code.) @
	go vet ./...


build-uploader: fmt vet
	go build -o output/images-uploader cmd/images/uploader/main.go

run-uploader: fmt vet
	go run cmd/images/uploader/main.go

build-vault: fmt vet
	go build -o output/images-uploader cmd/images/uploader/main.go

run-vault: fmt vet
	go run cmd/images/uploader/main.go	

update-codegen: ## generetor clientset informer inderx code
	./hack/update-codegen.sh

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

# Generate manifests for CRDs
manifests: controller-gen
	$(CONTROLLER_GEN) crd paths="./pkg/apis/..." output:crd:artifacts:config=config/crds

build-middleware: fmt vet controller-gen manifests
	go build -o output/middleware-operator cmd/middleware/main.go

run-middleware: fmt vet
	go run cmd/middleware/main.go

build-sysevent: fmt vet controller-gen manifests
	go build -o output/sys-event cmd/sys-event/main.go

run-sysevent: fmt vet
	go run cmd/sys-event/main.go

SHELL := bash

# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 0.0.1

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GINKGO ?= $(LOCALBIN)/ginkgo

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.14.0
GOTOOLCHAIN_VERSION ?= go1.21.0
GOLANGCI_VERSION ?= v1.64.8

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.29

# Number of CPUs to be allocacted for testing
PROCS?=$(shell expr $(shell nproc --ignore 2) / 2)
PROC_CMD = --procs $(PROCS)

.PHONY: all
all: build

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: fmt
fmt: gowork ## Run go fmt against code.
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		pushd ./$$mod ; \
		go fmt "./..." || exit 1 ; \
		popd ; \
	done

.PHONY: vet
vet: gowork ## Run go vet against code.
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		pushd ./$$mod ; \
		go vet ./... || exit 1 ; \
		popd ; \
	done

.PHONY: test
test: gowork generate fmt vet envtest ginkgo ## Run tests.
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		pushd ./$$mod ; \
		if [ -f test/functional/suite_test.go ]; then \
			KUBEBUILDER_ASSETS="$(shell $(ENVTEST) -v debug --bin-dir $(LOCALBIN) use $(ENVTEST_K8S_VERSION) -p path)" $(GINKGO) --trace --cover --coverprofile cover.out --covermode=atomic ${PROC_CMD} $(GINKGO_ARGS) ./test/... || exit 1; \
		fi; \
		KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... --cover --coverprofile cover.out --covermode=atomic || exit 1; \
		popd ; \
	done

##@ Build

.PHONY: build
build: fmt vet ## Build a test lib-common binary.
	go build -o lib-common

CI_TOOLS_REPO := https://github.com/openstack-k8s-operators/openstack-k8s-operators-ci
CI_TOOLS_REPO_DIR = $(shell pwd)/CI_TOOLS_REPO
.PHONY: get-ci-tools
get-ci-tools: ## Retrieve CI tools repo for running tests
	if [ -d  "$(CI_TOOLS_REPO_DIR)" ]; then \
		echo "Ci tools exists"; \
		pushd "$(CI_TOOLS_REPO_DIR)"; \
		git pull --rebase; \
		popd; \
	else \
		git clone $(CI_TOOLS_REPO) "$(CI_TOOLS_REPO_DIR)"; \
	fi

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@c7e1dc9b

.PHONY: ginkgo
ginkgo: $(GINKGO) ## Download ginkgo locally if necessary.
$(GINKGO): $(LOCALBIN)
	test -s $(LOCALBIN)/ginkgo || GOBIN=$(LOCALBIN) go install github.com/onsi/ginkgo/v2/ginkgo

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./$$mod/..." ; \
	done

.PHONY: gofmt
gofmt: get-ci-tools ## Run go fmt via ci-tools script against code
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		GOWORK=off $(CI_TOOLS_REPO_DIR)/test-runner/gofmt.sh ./$$mod || exit 1 ; \
	done

.PHONY: govet
govet: get-ci-tools ## Run go vet via ci-tools script against code
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		GOWORK=off $(CI_TOOLS_REPO_DIR)/test-runner/govet.sh ./$$mod || exit 1 ; \
	done

.PHONY: gotest
gotest: get-ci-tools envtest ## Run go test via ci-tools script against code
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" GOWORK=off $(CI_TOOLS_REPO_DIR)/test-runner/gotest.sh ./$$mod || exit 1 ; \
	done

.PHONY: golangci
golangci: get-ci-tools ## Run golangci-lint test via ci-tools script against code
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		GOLANGCI_TAG=$(GOLANGCI_VERSION) GOWORK=off $(CI_TOOLS_REPO_DIR)/test-runner/golangci.sh ./$$mod || exit 1 ; \
	done

.PHONY: golint
golint: get-ci-tools ## Run go lint via ci-tools script against code
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		PATH=$(GOBIN):$(PATH); GOWORK=off $(CI_TOOLS_REPO_DIR)/test-runner/golint.sh $$mod || exit 1; \
	done

.PHONY: gowork
gowork: ## Initiate go work
	test -f go.work || GOTOOLCHAIN=$(GOTOOLCHAIN_VERSION) go work init
	go work use -r modules
	go work sync

.PHONY: tidy ## Run go tidy sequentially on all modules
tidy:
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		set -x; \
		pushd ./$$mod ; \
		go mod tidy; \
		popd; \
	done

.PHONY: operator-lint
operator-lint: gowork ## Runs operator-lint
	GOBIN=$(LOCALBIN) go install github.com/gibizer/operator-lint@v0.3.0
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		set -x ; \
	        pushd ./$$mod ; \
        	go vet -vettool=$(LOCALBIN)/operator-lint ./... || exit 1 ; \
	        popd ; \
	done

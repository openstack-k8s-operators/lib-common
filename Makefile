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

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= v0.9.2

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.24


.PHONY: all
all: build

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
test: gowork generate fmt vet envtest ## Run tests.
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		pushd ./$$mod ; \
		KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) -p path)" go test ./... -coverprofile cover.out || exit 1; \
		popd ; \
	done
##@ Build

.PHONY: build
build: fmt vet ## Build a test lib-common binary.
	go build -o lib-common

# CI tools repo for running tests
CI_TOOLS_REPO := https://github.com/openstack-k8s-operators/openstack-k8s-operators-ci
CI_TOOLS_REPO_DIR = $(shell pwd)/CI_TOOLS_REPO
.PHONY: get-ci-tools
get-ci-tools:
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
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./$$mod/..." ; \
	done

# Run go fmt via ci-tools script against code
.PHONY: gofmt
gofmt: get-ci-tools
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		GOWORK=off $(CI_TOOLS_REPO_DIR)/test-runner/gofmt.sh ./$$mod || exit 1 ; \
	done

# Run go vet via ci-tools script against code
.PHONY: govet
govet: get-ci-tools
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		GOWORK=off $(CI_TOOLS_REPO_DIR)/test-runner/govet.sh ./$$mod || exit 1 ; \
	done

# Run go test via ci-tools script against code
.PHONY: gotest
gotest: get-ci-tools
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		GOWORK=off $(CI_TOOLS_REPO_DIR)/test-runner/gotest.sh ./$$mod || exit 1 ; \
	done

# Run golangci-lint test via ci-tools script against code
.PHONY: golangci
golangci: get-ci-tools
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		GOWORK=off $(CI_TOOLS_REPO_DIR)/test-runner/golangci.sh ./$$mod || exit 1 ; \
	done

# Run go lint via ci-tools script against code
.PHONY: golint
golint: get-ci-tools
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		PATH=$(GOBIN):$(PATH); GOWORK=off $(CI_TOOLS_REPO_DIR)/test-runner/golint.sh $$mod || exit 1; \
	done

.PHONY: gowork
gowork:
	test -f go.work || go work init
	for mod in $(shell find modules -maxdepth 1 -mindepth 1 -type d); do go work use $$mod; done
	go work sync

.PHONY: operator-lint
operator-lint: gowork ## Runs operator-lint
	GOBIN=$(LOCALBIN) go install github.com/gibizer/operator-lint@v0.1.0
	for mod in $(shell find modules/ -maxdepth 1 -mindepth 1 -type d); do \
		set -x ; \
		if [ $$mod == "modules/archive" ]; then continue; fi ; \
        pushd ./$$mod ; \
        go vet -vettool=$(LOCALBIN)/operator-lint ./... || exit 1 ; \
        popd ; \
    done

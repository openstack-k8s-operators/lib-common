# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION ?= 0.0.1

.PHONY: all
all: build

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

##@ Build

.PHONY: build
build: fmt vet ## Build a test lib-common binary.
	go build -o lib-common

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


# Run go fmt against code
gofmt: get-ci-tools
	$(CI_TOOLS_REPO_DIR)/test-runner/gofmt.sh

# Run go vet against code
govet: get-ci-tools
	$(CI_TOOLS_REPO_DIR)/test-runner/govet.sh

# Run go test against code
gotest: get-ci-tools
	$(CI_TOOLS_REPO_DIR)/test-runner/gotest.sh

# Run golangci-lint test against code
golangci: get-ci-tools
	$(CI_TOOLS_REPO_DIR)/test-runner/golangci.sh

# Run go lint against code
golint: get-ci-tools
	PATH=$(GOBIN):$(PATH); $(CI_TOOLS_REPO_DIR)/test-runner/golint.sh

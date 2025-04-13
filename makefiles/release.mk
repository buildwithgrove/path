# Makefile for Path project
# Supports building, testing, releasing, and CI workflow testing

#############################################
##          Configuration variables        ##
#############################################
VERSION ?= $(shell git describe --tags --always)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BINARY_NAME ?= path
# Supported build platforms
PLATFORMS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64
# Build flags
LDFLAGS := -ldflags="-X main.Version=$(VERSION) -X main.Date=$(DATE)"
# Path to GitHub Actions workflows
GH_WORKFLOWS := .github/workflows
# Output directories
RELEASE_DIR := release
BIN_DIR := bin
# Architecture detection for M-series Macs
ARCH := $(shell uname -m)
ifeq ($(ARCH),arm64)
  # Check if running on macOS
  ifeq ($(shell uname),Darwin)
    ACT_ARCH_FLAG := --container-architecture linux/amd64
  endif
endif

#############################################
##             Build commands              ##
#############################################

.PHONY: path_build
path_build: ## Build the binary for local development
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd

.PHONY: path_release
path_release: ## Build release binaries for all supported platforms
	@mkdir -p $(RELEASE_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d/ -f1); \
		arch=$$(echo $$platform | cut -d/ -f2); \
		echo "Building for $$os/$$arch..."; \
		output_name="$(BINARY_NAME)_$(VERSION)_$${os}_$${arch}"; \
		GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $(RELEASE_DIR)/$$output_name ./cmd; \
		tar -czf $(RELEASE_DIR)/$$output_name.tar.gz -C $(RELEASE_DIR) $$output_name; \
		rm $(RELEASE_DIR)/$$output_name; \
	done
	@echo "✅ Release binaries built in the '$(RELEASE_DIR)' directory"

#############################################
##          Versioning commands            ##
#############################################

.PHONY: path_release_tag
path_release_tag: ## Tag a new release with specified type (bug, minor, major)
	@echo "What type of release? [bug/minor/major]"
	@read TYPE && \
	LATEST_TAG=$$(git tag --sort=-v:refname | head -n 1) && \
	if [ -z "$$LATEST_TAG" ]; then \
		echo "No existing tags found. Creating initial v0.1.0 tag."; \
		NEW_TAG="v0.1.0"; \
	elif [ "$$TYPE" = "bug" ]; then \
		NEW_TAG=$$(echo $$LATEST_TAG | awk -F. -v OFS=. '{ $$NF = sprintf("%d", $$NF + 1); print }'); \
	elif [ "$$TYPE" = "minor" ]; then \
		NEW_TAG=$$(echo $$LATEST_TAG | awk -F. '{$$2 += 1; $$3 = 0; print $$1 "." $$2 "." $$3}'); \
	elif [ "$$TYPE" = "major" ]; then \
		NEW_TAG=$$(echo $$LATEST_TAG | awk -F. '{$$1 = substr($$1,2) + 0; $$2 = 0; $$3 = 0; print "v" $$1 "." $$2 "." $$3}'); \
	else \
		echo "❌ Invalid type. Use 'bug', 'minor', or 'major'."; \
		exit 1; \
	fi && \
	git tag $$NEW_TAG && \
	echo "✅ New version tagged: $$NEW_TAG" && \
	echo "Run 'git push origin $$NEW_TAG' to trigger the release workflow"

#############################################
##       CI/CD Workflow Testing            ##
#############################################

.PHONY: check_act
# Internal helper: Check if act is installed
check_act:
	@if ! command -v act >/dev/null 2>&1; then \
		echo "❌ Please install act first with 'make install_act'"; \
		exit 1; \
	fi;

.PHONY: check_secrets
# Internal helper: Check if .secrets file exists with valid GITHUB_TOKEN
check_secrets:
	@if [ ! -f .secrets ]; then \
		echo "❌ .secrets file not found!"; \
		echo "Please create a .secrets file with your GitHub token:"; \
		echo "GITHUB_TOKEN=your_github_token"; \
		exit 1; \
	fi
	@if ! grep -q "GITHUB_TOKEN=" .secrets; then \
		echo "❌ GITHUB_TOKEN not found in .secrets file!"; \
		echo "Please add GITHUB_TOKEN to your .secrets file:"; \
		echo "GITHUB_TOKEN=your_github_token"; \
		echo "You can create a token at: https://github.com/settings/tokens"; \
		exit 1; \
	fi
	@if grep -q "GITHUB_TOKEN=$$" .secrets || grep -q "GITHUB_TOKEN=\"\"" .secrets || grep -q "GITHUB_TOKEN=''" .secrets; then \
		echo "❌ GITHUB_TOKEN is empty in .secrets file!"; \
		echo "Please set a valid GitHub token:"; \
		echo "GITHUB_TOKEN=your_github_token"; \
		echo "You can create a token at: https://github.com/settings/tokens"; \
		exit 1; \
	fi

.PHONY: install_act
install_act: ## Install act for local GitHub Actions testing
	@echo "Installing act..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		brew install act; \
	else \
		curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash; \
	fi
	@echo "✅ act installed successfully"

.PHONY: workflow_test_build_and_push
workflow_test_build_and_push: check_act check_secrets ## Test the build and push GitHub workflow
	@echo "Testing build and push workflow..."
	@act -W $(GH_WORKFLOWS)/build-and-push.yml workflow_dispatch $(ACT_ARCH_FLAG) -v --secret-file .secrets

.PHONY: workflow_test_release
workflow_test_release: check_act check_secrets ## Test the release GitHub workflow
	@echo "Testing release workflow with custom tag 'test-release'..."
	@act -W $(GH_WORKFLOWS)/release-artifacts.yml workflow_dispatch -P custom_tag=test-release $(ACT_ARCH_FLAG) -v --secret-file .secrets

.PHONY: workflow_test_all
workflow_test_all: check_act check_secrets ## Test all GitHub Actions workflows
	@echo "Testing all workflows..."
	@echo "1. Testing build and push workflow..."
	@act -W $(GH_WORKFLOWS)/build-and-push.yml workflow_dispatch $(ACT_ARCH_FLAG) -v
	@echo "2. Testing release workflow..."
	@act -W $(GH_WORKFLOWS)/release-artifacts.yml workflow_dispatch -P custom_tag=test-release $(ACT_ARCH_FLAG) -v --secret-file .secrets
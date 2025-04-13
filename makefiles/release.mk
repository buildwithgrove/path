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

#############################################
##             Build commands              ##
#############################################
.PHONY: build
build: ## Build the binary for local development
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd

.PHONY: release
release: ## Build release binaries for all supported platforms
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
.PHONY: release_tag
release_tag: ## Tag a new release with specified type (bug, minor, major)
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

.PHONY: install_act
install_act: ## Install act for local GitHub Actions testing
	@echo "Installing act..."
	@if [ "$(uname)" = "Darwin" ]; then \
		brew install act; \
	else \
		curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash; \
	fi
	@echo "✅ act installed successfully"

.PHONY: test_workflow_ci
test_workflow_ci: check_act ## Test the CI workflow (Build and push to ghcr.io)
	@echo "Testing CI workflow (Build and push to ghcr.io)..."
	@act -W $(GH_WORKFLOWS)/build-and-push.yml workflow_dispatch -v

.PHONY: test_workflow_release
test_workflow_release: check_act ## Test the release workflow
	@echo "Testing release workflow with custom tag 'test-release'..."
	@act -W $(GH_WORKFLOWS)/release-artifacts.yml workflow_dispatch -P custom_tag=test-release -v

.PHONY: test_workflows_all
test_workflows_all: check_act ## Test all GitHub Actions workflows
	@echo "Testing all workflows..."
	@echo "1. Testing CI workflow..."
	@act -W $(GH_WORKFLOWS)/build-and-push.yml workflow_dispatch -v
	@echo "2. Testing release workflow..."
	@act -W $(GH_WORKFLOWS)/release-artifacts.yml workflow_dispatch -P custom_tag=test-release -v

# Set default target to help
.DEFAULT_GOAL := help
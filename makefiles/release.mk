
###########################
###   Release Helpers   ###
###########################

# Sync local tags to remote:
# git fetch --prune --prune-tags origin

.PHONY: release_tag_local_testing
release_tag_local_testing: ## Tag a new local testing release (e.g. v1.0.1 -> v1.0.2-test1, v1.0.2-test1 -> v1.0.2-test2)
	@LATEST_TAG=$$(git tag --sort=-v:refname | head -n 1 | xargs); \
	if [ -z "$$LATEST_TAG" ]; then \
	  NEW_TAG=v0.1.0-test1; \
	else \
	  if echo "$$LATEST_TAG" | grep -q -- '-test'; then \
	    BASE_TAG=$$(echo "$$LATEST_TAG" | sed 's/-test[0-9]*//'); \
	    LAST_TEST_NUM=$$(echo "$$LATEST_TAG" | sed -E 's/.*-test([0-9]+)/\1/'); \
	    NEXT_TEST_NUM=$$(($$LAST_TEST_NUM + 1)); \
	    NEW_TAG=$${BASE_TAG}-test$${NEXT_TEST_NUM}; \
	  else \
	    BASE_TAG=$$(echo "$$LATEST_TAG" | awk -F. -v OFS=. '{$$NF = sprintf("%d", $$NF + 1); print}'); \
	    NEW_TAG=$${BASE_TAG}-test1; \
	  fi; \
	fi; \
	git tag $$NEW_TAG; \
	echo "New local testing version tagged: $$NEW_TAG"; \
	echo "Run the following commands to push the new tag:"; \
	echo "  git push origin $$NEW_TAG"; \
	echo "And draft a new release at https://github.com/buildwithgrove/path/releases/new";


.PHONY: release_tag_dev
release_tag_dev: ## Tag a new dev release (e.g. v1.0.1 -> v1.0.1-dev1, v1.0.1-dev1 -> v1.0.1-dev2)
	@$(eval LATEST_TAG=$(shell git tag --sort=-v:refname | head -n 1))
	@$(eval BASE_VERSION=$(shell echo $(LATEST_TAG) | sed 's/-dev[0-9]*$$//' ))
	@$(eval EXISTING_DEV_TAGS=$(shell git tag --sort=-v:refname | grep "^$(BASE_VERSION)-dev[0-9]*$$" | head -n 1))
	@if [ -z "$(EXISTING_DEV_TAGS)" ]; then \
		NEW_TAG="$(BASE_VERSION)-dev1"; \
	else \
		DEV_NUM=$$(echo $(EXISTING_DEV_TAGS) | sed 's/.*-dev\([0-9]*\)$$/\1/'); \
		NEW_DEV_NUM=$$((DEV_NUM + 1)); \
		NEW_TAG="$(BASE_VERSION)-dev$$NEW_DEV_NUM"; \
	fi; \
	git tag $$NEW_TAG; \
	echo "########"; \
	echo "New dev version tagged: $$NEW_TAG"; \
	echo ""; \
	echo "Next, do the following:"; \
	echo ""; \
	echo "1. Run the following commands to push the new tag:"; \
	echo "   git push origin $$NEW_TAG"; \
	echo ""; \
	echo "2. And draft a new release at https://github.com/buildwithgrove/path/releases/new"; \
	echo ""; \
	echo "If you need to delete a tag, run:"; \
	echo "  git tag -d $$NEW_TAG"; \
	echo ""; \
	echo "If you need to delete a tag remotely, run:"; \
	echo "  git push origin --delete $$NEW_TAG"; \
	echo ""; \
	echo "########"

.PHONY: release_tag_bug_fix
release_tag_bug_fix: ## Tag a new bug fix release (e.g. v1.0.1 -> v1.0.2)
	@$(eval LATEST_TAG=$(shell git tag --sort=-v:refname | head -n 1))
	@LATEST_TAG="$(LATEST_TAG)"; \
	NEW_TAG=$$(echo $$LATEST_TAG | awk -F. -v OFS=. '{ $$NF = sprintf("%d", $$NF + 1); print }'); \
	git tag $$NEW_TAG; \
	echo "########"; \
	echo "New bug fix version tagged: $$NEW_TAG"; \
	echo ""; \
	echo "Next, do the following:"; \
	echo ""; \
	echo "1. Run the following commands to push the new tag:"; \
	echo "   git push origin $$NEW_TAG"; \
	echo ""; \
	echo "2. And draft a new release at https://github.com/buildwithgrove/path/releases/new"; \
	echo ""; \
	echo "If you need to delete a tag, run:"; \
	echo "  git tag -d $$NEW_TAG"; \
	echo ""; \
	echo "If you need to delete a tag remotely, run:"; \
	echo "  git push origin --delete $$NEW_TAG"; \
	echo ""; \
	echo "########"


.PHONY: release_tag_minor_release
release_tag_minor_release: ## Tag a new minor release (e.g. v1.0.0 -> v1.1.0)
	@$(eval LATEST_TAG=$(shell git tag --sort=-v:refname | head -n 1))
	@LATEST_TAG="$(LATEST_TAG)"; \
	NEW_TAG=$$(echo $$LATEST_TAG | awk -F. '{$$2 += 1; $$3 = 0; print $$1 "." $$2 "." $$3}'); \
	git tag $$NEW_TAG; \
	echo "########"; \
	echo "New minor release version tagged: $$NEW_TAG"; \
	echo ""; \
	echo "Next, do the following:"; \
	echo ""; \
	echo "1. Run the following commands to push the new tag:"; \
	echo "   git push origin $$NEW_TAG"; \
	echo ""; \
	echo "2. And draft a new release at https://github.com/buildwithgrove/path/releases/new"; \
	echo ""; \
	echo "If you need to delete a tag, run:"; \
	echo "  git tag -d $$NEW_TAG"; \
	echo ""; \
	echo "If you need to delete a tag remotely, run:"; \
	echo "  git push origin --delete $$NEW_TAG"; \
	echo ""; \
	echo "########";

#############################
### Binary Build Targets  ###
#############################

# Define the release directory
RELEASE_DIR ?= ./release

# Define the architectures we want to build for
RELEASE_PLATFORMS := linux/amd64 linux/arm64

# Version information (can be overridden)
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse HEAD)
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go build flags
LDFLAGS := -s -w \
	-X main.Version=$(VERSION) \
	-X main.Commit=$(COMMIT) \
	-X main.BuildDate=$(BUILD_DATE)

.PHONY: release_build_cross
release_build_cross: ## Build binaries for multiple platforms
	@echo "Building binaries for multiple platforms..."
	@mkdir -p $(RELEASE_DIR)
	@for platform in $(RELEASE_PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d/ -f1); \
		GOARCH=$$(echo $$platform | cut -d/ -f2); \
		output=$(RELEASE_DIR)/path-$$GOOS-$$GOARCH; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags="$(LDFLAGS)" -o $$output ./cmd; \
		if [ $$? -eq 0 ]; then \
			echo "✓ Built $$output"; \
		else \
			echo "✗ Failed to build for $$GOOS/$$GOARCH"; \
			exit 1; \
		fi; \
	done
	@echo "All binaries built successfully!"

.PHONY: release_clean
release_clean: ## Clean up release artifacts
	@echo "Cleaning release directory..."
	@rm -rf $(RELEASE_DIR)

.PHONY: release_build_local
release_build_local: ## Build binary for current platform only
	@echo "Building binary for current platform..."
	@mkdir -p $(RELEASE_DIR)
	@CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(RELEASE_DIR)/path-local ./cmd
	@echo "✓ Built $(RELEASE_DIR)/path-local"

.PHONY: build_ghcr_image_current_branch
build_ghcr_image_current_branch: ## Trigger the main-build workflow using the current branch to push an image to ghcr.io/buildwithgrove/path
	@echo "Triggering main-build workflow for current branch..."
	@BRANCH=$$(git rev-parse --abbrev-ref HEAD) && \
	gh workflow run main-build.yml --ref $$BRANCH
	@echo "Workflow triggered for branch: ${CYAN} $$(git rev-parse --abbrev-ref HEAD)${RESET}"
	@echo "Check the workflow status at: ${BLUE}https://github.com/$(shell git config --get remote.origin.url | sed 's/.*github.com[:/]\([^/]*\/[^.]*\).*/\1/')/actions/workflows/main-build.yml${RESET}"

.PHONY: release_ghcr_image_current_branch
release_ghcr_image_current_branch: build_ghcr_image_current_branch ## Trigger the main-build workflow using the current branch to push an image to ghcr.io/buildwithgrove/path
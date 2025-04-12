###########################
###   Release Helpers   ###
###########################

VERSION ?= $(shell git describe --tags --always)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
PLATFORMS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64
BINARY_NAME ?= path

# List tags: git tag
# Delete tag locally: git tag -d v1.2.3
# Delete tag remotely: git push --delete origin v1.2.3

.PHONY: install_act
install_act: ## Install act for local GitHub Actions testing
	@echo "Installing act..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		brew install act; \
	elif [ "$$(uname)" = "Linux" ]; then \
		curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash; \
	else \
		echo "Please install act manually: https://github.com/nektos/act#installation"; \
	fi

.PHONY: release_tag_bug_fix
release_tag_bug_fix: ## Tag a new bug fix release (e.g. v1.0.1 -> v1.0.2)
	@$(eval LATEST_TAG=$(shell git tag --sort=-v:refname | head -n 1))
	@$(eval NEW_TAG=$(shell echo $(LATEST_TAG) | awk -F. -v OFS=. '{ $$NF = sprintf("%d", $$NF + 1); print }'))
	@git tag $(NEW_TAG)
	@echo "New bug fix version tagged: $(NEW_TAG)"
	@echo "Run the following commands to push the new tag:"
	@echo "  git push origin $(NEW_TAG)"
	@echo "This will trigger the release workflow automatically"

.PHONY: release_tag_minor_release
release_tag_minor_release: ## Tag a new minor release (e.g. v1.0.0 -> v1.1.0)
	@$(eval LATEST_TAG=$(shell git tag --sort=-v:refname | head -n 1))
	@$(eval NEW_TAG=$(shell echo $(LATEST_TAG) | awk -F. '{$$2 += 1; $$3 = 0; print $$1 "." $$2 "." $$3}'))
	@git tag $(NEW_TAG)
	@echo "New minor release version tagged: $(NEW_TAG)"
	@echo "Run the following commands to push the new tag:"
	@echo "  git push origin $(NEW_TAG)"
	@echo "This will trigger the release workflow automatically"

.PHONY: path_build
path_build: ## Build the PATH binary for local development
	go build -ldflags="-X main.Version=$(VERSION) -X main.Date=$(DATE)" -o bin/$(BINARY_NAME) ./cmd

.PHONY: path_build_release
path_build_release: ## Build release binaries for all supported platforms
	@mkdir -p release
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d/ -f1); \
		arch=$$(echo $$platform | cut -d/ -f2); \
		echo "Building for $$os/$$arch..."; \
		output_name="$(BINARY_NAME)_$(VERSION)_$${os}_$${arch}"; \
		if [ "$$os" = "windows" ]; then output_name="$$output_name.exe"; fi; \
		GOOS=$$os GOARCH=$$arch go build -ldflags="-X main.Version=$(VERSION) -X main.Date=$(DATE)" -o release/$$output_name ./cmd; \
		if [ "$$os" != "windows" ]; then \
			tar -czf release/$$output_name.tar.gz -C release $$output_name; \
			rm release/$$output_name; \
		else \
			zip -j release/$$output_name.zip release/$$output_name; \
			rm release/$$output_name.exe; \
		fi; \
	done
	@echo "Release binaries built in the 'release' directory"

.PHONY: path_release_publish
path_release_publish: release_tag_bug_fix ## Tag and prepare for release
	@echo "To publish the release, push the tag with: git push origin $$(git tag --sort=-v:refname | head -n 1)"

.PHONY: path_release_publish_test
path_release_publish_test: ## Test the release workflow locally using act
	@command -v act >/dev/null 2>&1 || { echo "Please install act first. See: https://github.com/nektos/act"; exit 1; }
	@echo "Testing release workflow specifically..."
	@act -W .github/workflows/release-artifacts.yml workflow_dispatch -P custom_tag=test-release -v

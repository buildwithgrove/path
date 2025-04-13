###########################
###   Release Helpers   ###
###########################

VERSION ?= $(shell git describe --tags --always)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
PLATFORMS ?= linux/amd64 linux/arm64 darwin/amd64 darwin/arm64
BINARY_NAME ?= path

.PHONY: install_act
install_act: ## Install act for local GitHub Actions testing
	@echo "Installing act..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		brew install act; \
	else \
		curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash; \
	fi

.PHONY: release_tag
release_tag: ## Tag a new release with specified type (bug, minor, major)
	@echo "What type of release? [bug/minor/major]"
	@read TYPE && \
	LATEST_TAG=$$(git tag --sort=-v:refname | head -n 1) && \
	if [ "$$TYPE" = "bug" ]; then \
		NEW_TAG=$$(echo $$LATEST_TAG | awk -F. -v OFS=. '{ $$NF = sprintf("%d", $$NF + 1); print }'); \
	elif [ "$$TYPE" = "minor" ]; then \
		NEW_TAG=$$(echo $$LATEST_TAG | awk -F. '{$$2 += 1; $$3 = 0; print $$1 "." $$2 "." $$3}'); \
	elif [ "$$TYPE" = "major" ]; then \
		NEW_TAG=$$(echo $$LATEST_TAG | awk -F. '{$$1 = substr($$1,2) + 0; $$2 = 0; $$3 = 0; print "v" $$1 "." $$2 "." $$3}'); \
	else \
		echo "Invalid type. Use 'bug', 'minor', or 'major'."; \
		exit 1; \
	fi && \
	git tag $$NEW_TAG && \
	echo "New version tagged: $$NEW_TAG" && \
	echo "Run 'git push origin $$NEW_TAG' to trigger the release workflow"

.PHONY: build
build: ## Build the binary for local development
	go build -ldflags="-X main.Version=$(VERSION) -X main.Date=$(DATE)" -o bin/$(BINARY_NAME) ./cmd

.PHONY: release
release: ## Build release binaries for all supported platforms
	@mkdir -p release
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d/ -f1); \
		arch=$$(echo $$platform | cut -d/ -f2); \
		echo "Building for $$os/$$arch..."; \
		output_name="$(BINARY_NAME)_$(VERSION)_$${os}_$${arch}"; \
		GOOS=$$os GOARCH=$$arch go build -ldflags="-X main.Version=$(VERSION) -X main.Date=$(DATE)" -o release/$$output_name ./cmd; \
		tar -czf release/$$output_name.tar.gz -C release $$output_name; \
		rm release/$$output_name; \
	done
	@echo "Release binaries built in the 'release' directory"

.PHONY: test_release_workflow
test_release_workflow: ## Test the release workflow locally using act
	@command -v act >/dev/null 2>&1 || { echo "Please install act first with 'make install_act'"; exit 1; }
	@act -W .github/workflows/release-artifacts.yml workflow_dispatch -P custom_tag=test-release -v
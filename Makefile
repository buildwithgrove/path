########################
### Makefile Helpers ###
########################

# TODO(@olshansk): Remove "Shannon" and just use "Pocket".

.PHONY: help
.DEFAULT_GOAL := help
help: ## Prints all the targets in all the Makefiles
	@echo ""
	@echo "$(BOLD)$(CYAN)🌐 PATH (Path API & Toolkit Harness) Makefile Targets$(RESET)"
	@echo ""
	@echo "$(BOLD)=== 📋 Information & Discovery ===$(RESET)"
	@grep -h -E '^help:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 🔨 Build & Run ===$(RESET)"
	@grep -h -E '^path_(build|run):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== ⚙️ Configuration ===$(RESET)"
	@grep -h -E '^config.*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 🛠️ Development Environment ===$(RESET)"
	@grep -h -E '^(path_up|path_down|install_tools.*|localnet_.*):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 🚀 Load Testing ===$(RESET)"
	@grep -h -E '^load_test.*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 🧪 Testing ===$(RESET)"
	@grep -h -E '^(test_unit|test_all|go_lint):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@grep -h -E '^e2e_test.*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== ⚡ Benchmarking ===$(RESET)"
	@grep -h -E '^bench.*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== ✋ Manual Testing ===$(RESET)"
	@grep -h -E '^(get_disqualified_endpoints|grove_get_disqualified_endpoints|shannon_preliminary_services_test_help|shannon_preliminary_services_test|source_shannon_preliminary_services_helpers):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 🗄️ Portal Database ===$(RESET)"
	@grep -h -E '^portal_db.*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 📦 Protocol Buffers ===$(RESET)"
	@grep -h -E '^proto.*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 🚢 Release Management ===$(RESET)"
	@grep -h -E '^release_(tag_.*|build_cross|build_local|clean):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@grep -h -E '^(release_ghcr_image_current_branch|release_build_and_push_localnet_image):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 🔧 Utilities ===$(RESET)"
	@echo ""
	@echo "$(BOLD)=== 📚 Documentation ===$(RESET)"
	@grep -h -E '^(go_docs|docusaurus.*|gen_.*_docs):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 🔍 Request Testing ===$(RESET)"
	@grep -h -E '^test_(request|healthz|disqualified|load).*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== ⚡ Benchmarking ===$(RESET)"
	@grep -h -E '^bench_.*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(BOLD)=== 🤖 AI ===$(RESET)"
	@grep -h -E '^claudesync.*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "$(CYAN)%-40s$(RESET) %s\n", $$1, $$2}'
	@echo ""

#############################
#### PATH Build Targets   ###
#############################

# tl;dr Quick testing & debugging of PATH as a standalone
# This section is intended to just build and run the PATH binary.
# It mimics an E2E real environment.

.PHONY: path_build
path_build: ## Build the path binary locally (does not run anything)
	go build -o bin/path ./cmd

# The PATH config value can be set via the CONFIG_PATH env variable and defaults to ./local/path/.config.yaml
CONFIG_PATH ?= ./local/path/.config.yaml

.PHONY: check_path_config
## Verify that path configuration file exists
check_path_config:
	@if [ -z "$(CONFIG_PATH)" ]; then \
		echo "################################################################"; \
		echo "Error: CONFIG_PATH is not set."; \
		echo ""; \
		echo "Set CONFIG_PATH to your config file, e.g.:"; \
		echo "  export CONFIG_PATH=./local/path/.config.yaml"; \
		echo "Or initialize using:"; \
		echo "  make config_prepare_shannon_e2e"; \
		echo "################################################################"; \
		exit 1; \
	fi

.PHONY: path_run
path_run: path_build check_path_config ## Run the path binary as a standalone binary
	(cd bin; ./path -config ../${CONFIG_PATH})

###############################
###    Makefile imports     ###
###############################

include ./makefiles/colors.mk
include ./makefiles/configs.mk
include ./makefiles/configs_shannon.mk
include ./makefiles/deps.mk
include ./makefiles/devtools.mk
include ./makefiles/docs.mk
include ./makefiles/localnet.mk
include ./makefiles/portal-db.mk
include ./makefiles/test.mk
include ./makefiles/bench.mk
include ./makefiles/test_requests.mk
include ./makefiles/test_load.mk
include ./makefiles/proto.mk
include ./makefiles/debug.mk
include ./makefiles/claude.mk
include ./makefiles/release.mk
include ./makefiles/helpers.mk

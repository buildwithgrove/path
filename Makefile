########################
### Makefile Helpers ###
########################

# TODO(@olshansk): Remove "Shannon" and just use "Pocket".

.PHONY: list
list: ## List all make targets
	@${MAKE} -pRrn : -f $(MAKEFILE_LIST) 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | sort

.PHONY: help
.DEFAULT_GOAL := help
help: ## Prints all the targets in all the Makefiles
	@grep -h -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-60s\033[0m %s\n", $$1, $$2}'

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

include ./makefiles/configs.mk
include ./makefiles/configs_shannon.mk
include ./makefiles/deps.mk
include ./makefiles/devtools.mk
include ./makefiles/docs.mk
include ./makefiles/localnet.mk
include ./makefiles/portal-db.mk
include ./makefiles/test.mk
include ./makefiles/test_requests.mk
include ./makefiles/test_load.mk
include ./makefiles/proto.mk
include ./makefiles/debug.mk
include ./makefiles/claude.mk
include ./makefiles/release.mk
include ./makefiles/helpers.mk

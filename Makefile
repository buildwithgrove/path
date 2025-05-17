########################
### Makefile Helpers ###
########################

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

# The PATH config value can be overridden via the CONFIG_PATH env variable, which defaults to ../local/path/.config.yaml
# This path must be either an absolute path or relative to the location of the PATH binary in `bin`.
#
# Example usage:
# - absolute path
# 	make path_run CONFIG_PATH=/Users/greg/path/local/path/.config.yaml
# - relative path
# 	make path_run CONFIG_PATH=../local/path/.config.yaml
CONFIG_PATH ?= ../local/path/.config.yaml

.PHONY: path_run
path_run: path_build check_path_config ## Run the path binary as a standalone binary
	(cd bin; ./path -config ${CONFIG_PATH})

.PHONY: check_path_config
## Verify that path configuration file exists
check_path_config:
	@if [ ! -f ./local/path/.config.yaml ]; then \
		echo "################################################################"; \
   		echo "Error: Missing config file at ./local/path/.config.yaml"; \
   		echo ""; \
   		echo "Initialize using either:"; \
   		echo "  make shannon_prepare_e2e_config"; \
   		echo "  make morse_prepare_e2e_config "; \
   		echo "################################################################"; \
   		exit 1; \
   fi

.PHONY: check_docker
# Internal helper: Check if Docker is installed locally
check_docker:
	@if ! command -v docker >/dev/null 2>&1; then \
		echo "Docker is not installed. Make sure you review README.md before continuing"; \
		exit 1; \
	fi;

###############################
###    Makefile imports     ###
###############################

include ./makefiles/configs.mk
include ./makefiles/configs_morse.mk
include ./makefiles/configs_shannon.mk
include ./makefiles/deps.mk
include ./makefiles/docs.mk
include ./makefiles/localnet.mk
include ./makefiles/test.mk
include ./makefiles/test_requests.mk
include ./makefiles/proto.mk
include ./makefiles/debug.mk
include ./makefiles/claude.mk

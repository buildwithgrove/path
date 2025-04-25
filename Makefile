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

# The PATH config value can be set via the CONFIG_PATH env variable and defaults to ./local/path/.config.yaml
CONFIG_PATH ?= ../local/path/.config.yaml

.PHONY: path_run
path_run: path_build check_path_config ## Run the path binary as a standalone binary
	(cd bin; ./path -config ${CONFIG_PATH})

#################################
###  Local PATH make targets  ###
#################################

# tl;dr Mimic an E2E real environment.
# This section is intended to spin up and develop a full modular stack that includes
# PATH, Envoy Proxy, Rate Limiter, Auth Server, and any other dependencies.

.PHONY: path_up
path_up: check_path_config k8s_prepare_local_env ## Brings up local Tilt development environment which includes PATH and all related dependencies (using kind cluster)
	tilt up

.PHONY: path_down
path_down: ## Tears down local Tilt development environment which includes PATH and all related dependencies (using kind cluster)
	tilt down

.PHONY: path_help
path_help: ## Prints help commands if you cannot start path
	@echo "################################################################";
	@echo "If you're hitting issues running PATH, try running following commands:";
	@echo "	make path_down";
	@echo "	make k8s_cleanup_local_env";
	@echo "	make path_up";
	@echo "################################################################";


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

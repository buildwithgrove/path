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
check_path_config: ## Verify that path configuration file exists
	@if [ ! -f ./local/path/config/.config.yaml ]; then \
		echo "################################################################"; \
   		echo "Error: Missing config file at ./local/path/config/.config.yaml"; \
   		echo ""; \
   		echo "Initialize using either:"; \
   		echo "  make prepare_shannon_e2e_config"; \
   		echo "  make prepare_morse_e2e_config "; \
   		echo "################################################################"; \
   		exit 1; \
   fi

.PHONY: path_run
path_run: path_build check_path_config ## Run the path binary as a standalone binary
	(cd bin; ./path -config ../local/path/config/.config.yaml)

#################################
###  Local PATH make targets  ###
#################################

# tl;dr Mimic an E2E real environment.
# This section is intended to spin up and develop a full modular stack that includes
# PATH, Envoy Proxy, Rate Limiter, Auth Server, and any other dependencies.

.PHONY: path_up
path_up: localnet_up ## Brings up local Tilt development environment which includes PATH and all related dependencies (using kind cluster)
	MODE=path_with_auth $(MAKE) localnet_up

.PHONY: path_up_standalone
path_up_standalone: ## Brings up local Tilt development environment with PATH only
	MODE=path_only $(MAKE) path_up

.PHONY: path_down
path_down: dev_down ## Tears down local Tilt development environment which includes PATH and all related dependencies (using kind cluster)

###############################
###    Makefile imports     ###
###############################

include ./makefiles/configs.mk
include ./makefiles/deps.mk
include ./makefiles/docs.mk
include ./makefiles/envoy.mk
include ./makefiles/localnet.mk
include ./makefiles/morse_configs.mk
include ./makefiles/shannon_configs.mk
include ./makefiles/test.mk
include ./makefiles/test_requests.mk
include ./makefiles/proto.mk

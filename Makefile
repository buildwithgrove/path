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

# TODO_IMPROVE: add a make target to generate mocks for all the interfaces in the project

#############################
#### PATH Build Targets   ###
#############################

# tl;dr Quick testing & debugging of PATH as a standalone

# This section is intended to just build and run the PATH binary.
# It mimics an E2E real environment.

.PHONY: path_build
path_build: ## Build the path binary locally (does not run anything)
	go build -o bin/path ./cmd

.PHONY: path_run
path_run: path_build ## Run the path binary as a standalone binary
	@if [ ! -f ./bin/config/.config.yaml ]; then \
		echo "#########################################################################################"; \
		echo "### ./bin/config/.config.yaml does not exist, use ONE the following to initialize it: ###"; \
		echo "### A. make copy_shannon_config                                                       ###"; \
		echo "### B. make copy_morse_e2e_config                                                         ###"; \
		echo "#########################################################################################"; \
		exit; \
	fi; \
	(cd bin; ./path)

###############################
###  Localnet Make targets  ###
###############################

# tl;dr Mimic an E2E real environment.

# This section is intended to spin up and develop a full modular stack that includes
# PATH, Envoy Proxy, Rate Limiter, Auth Server, and any other dependencies.

.PHONY: localnet_up
localnet_up: config_shannon_localnet dev_up config_path_secrets ## Brings up local Tilt development environment which includes PATH and all related dependencies (using kind cluster)
	@tilt up

# NOTE: This is an intentional copy of localnet_up to enforce that the two are the same.
.PHONY: path_up
path_up: localnet_up ## Brings up local Tilt development environment which includes PATH and all related dependencies (using kind cluster)

.PHONY: path_up_standalone
path_up_standalone: ## Brings up local Tilt development environment with PATH only
	MODE=path_only $(MAKE) localnet_up

.PHONY: localnet_down
localnet_down: dev_down ## Tears down local Tilt development environment which includes PATH and all related dependencies (using kind cluster)

# NOTE: This is an intentional copy of localnet_down to enforce that the two are the same.
.PHONY: path_down
path_down: localnet_down ## Tears down local Tilt development environment which includes PATH and all related dependencies (using kind cluster)


###############################
### Generation Make Targets ###
###############################

# TODO_IMPROVE(@commoddity): update to use go:generate comments in the interface files and update this target

.PHONY: gen_proto
gen_proto: ## Generate the Go code from the gateway_endpoint.proto file
	protoc --go_out=./envoy/auth_server/proto --go-grpc_out=./envoy/auth_server/proto envoy/auth_server/proto/gateway_endpoint.proto

###############################
###    Makefile imports     ###
###############################

include ./makefiles/configs.mk
include ./makefiles/deps.mk
include ./makefiles/documentation.mk
include ./makefiles/envoy.mk
include ./makefiles/localnet.mk
include ./makefiles/morse_configs.mk
include ./makefiles/shannon_configs.mk
include ./makefiles/test.mk
include ./makefiles/test_requests.mk

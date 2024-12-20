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
		echo "### B. make copy_morse_config                                                         ###"; \
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

.PHONY: localnet_down
localnet_down: dev_down ## Tears down local Tilt development environment which includes PATH and all related dependencies (using kind cluster)

# NOTE: This is an intentional copy of localnet_down to enforce that the two are the same.
.PHONY: path_down
path_down: localnet_down ## Tears down local Tilt development environment which includes PATH and all related dependencies (using kind cluster)

#########################
### Test Make Targets ###
#########################

.PHONY: test_all ## Run all tests
test_all: test_unit test_auth_server test_e2e_shannon_relay test_e2e_morse_relay

.PHONY: test_unit
test_unit: ## Run all unit tests
	go test ./... -short -count=1

.PHONY: test_auth_server
test_auth_server: ## Run the auth server tests
	(cd envoy/auth_server && go test ./... -count=1)

.PHONY: test_e2e_shannon_relay_iterate
test_e2e_shannon_relay_iterate: ## Iterate on E2E shannon relay tests
	@echo "go build -o bin/path ./cmd"
	@echo "# Update ./bin/config/.config.yaml"
	@echo "./bin/path"
	@echo "curl http://anvil.localhost:3000/v1/abcd1234 -X POST -H \"Content-Type: application/json\" -d '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"eth_blockNumber\"}'"

.PHONY: test_e2e_shannon_relay
test_e2e_shannon_relay: shannon_e2e_config_warning ## Run an E2E Shannon relay test
	@echo "###############################################################################################################################################################"
	@echo "### README: If you are intending to iterate on E2E tests, stop this and run the following for instructions instead: 'make test_e2e_shannon_relay_iterate'. ###"
	@echo "###############################################################################################################################################################"
	go test -v ./e2e/... -tags=e2e -count=1 -run Test_ShannonRelay

.PHONY: test_e2e_morse_relay
test_e2e_morse_relay: morse_e2e_config_warning ## Run an E2E Morse relay test
	go test -v ./e2e/... -tags=e2e -count=1 -run Test_MorseRelay

###################################
### Shannon Config Make Targets ###
###################################

# TODO_MVP(@commoddity): Consolidate the copy_*_config targets into fewer targets once
# the config files are consolidated as well.

.PHONY: copy_shannon_config
copy_shannon_config: ## copies the example shannon configuration yaml file to .config.yaml file
	@if [ ! -f ./bin/config/.config.yaml ]; then \
		mkdir -p bin/config; \
		cp ./config/examples/config.shannon_example.yaml ./bin/config/.config.yaml; \
		echo "########################################################################"; \
		echo "### Created ./bin/config/.config.yaml                                ###"; \
		echo "### README: Please update the the following in .shannon.config.yaml: ###"; \
		echo "### 'gateway_private_key_hex' & 'owned_apps_private_keys_hex'.       ###"; \
		echo "########################################################################"; \
		echo "########################################################################"; \
	else \
		echo "##################################################################"; \
		echo "### ./bin/config/.config.yaml already exists, not overwriting. ###"; \
		echo "##################################################################"; \
	fi

.PHONY: shannon_e2e_config_warning
shannon_e2e_config_warning: ## Prints a warning if the shannon E2E config is not populated
	@if [ ! -f ./e2e/.shannon.config.yaml ]; then \
		echo "#########################################################################"; \
		echo "### Shannon E2E config not found, run: 'make copy_shannon_e2e_config' ###"; \
		echo "#########################################################################"; \
		exit; \
	fi

# If you are a Grove employee, search for this UUID in 1Password: 47k7kidj3y6nd3cghlakg76nlm
.PHONY: copy_shannon_e2e_config
copy_shannon_e2e_config: ## copies the example Shannon test configuration yaml file to .gitignored .shannon.config.yaml file
	@if [ ! -f ./e2e/.shannon.config.yaml ]; then \
		cp ./config/examples/config.shannon_example.yaml ./e2e/.shannon.config.yaml; \
		echo "########################################################################"; \
		echo "### Created ./e2e/.shannon.config.yaml                               ###"; \
		echo "### README: Please update the the following in .shannon.config.yaml: ###"; \
		echo "### 'gateway_private_key_hex' & 'owned_apps_private_keys_hex'.       ###"; \
		echo "########################################################################"; \
	else \
		echo "###################################################################"; \
		echo "### ./e2e/.shannon.config.yaml already exists, not overwriting. ###"; \
		echo "###################################################################"; \
	fi

.PHONY: config_shannon_localnet
config_shannon_localnet: ## Create a localnet config file to serve as a Shannon gateway
	@if [ ! -f ./local/path/config/.config.yaml ]; then \
		mkdir -p local/path/config; \
		cp ./config/examples/config.shannon_example.yaml  local/path/config/.config.yaml; \
		echo "########################################################################"; \
		echo "### Created ./local/path/config/.config.yaml for Shannon localnet.   ###"; \
		echo "### README: Please update the the following in .shannon.config.yaml: ###"; \
		echo "### 'gateway_private_key_hex' & 'owned_apps_private_keys_hex'.       ###"; \
		echo "########################################################################"; \
	else \
		echo "#########################################################################"; \
		echo "### ./local/path/config/.config.yaml already exists, not overwriting. ###"; \
		echo "#########################################################################"; \
	fi

###################################
### Morse Config Make Targets ###
###################################

.PHONY: copy_morse_config
copy_morse_config: ## copies the example morse configuration yaml file to .config.yaml file
	@if [ ! -f ./bin/config/.config.yaml ]; then \
		mkdir -p bin/config; \
		cp ./config/examples/config.morse_example.yaml ./bin/config/.config.yaml; \
		echo "######################################################################"; \
		echo "### Created ./bin/config/.config.yaml                              ###"; \
		echo "### README: Please update the the following in .morse.config.yaml: ###"; \
		echo "### 'url', 'relay_signing_key', & 'signed_aats'.                   ###"; \
		echo "######################################################################"; \
	else \
		echo "##################################################################"; \
		echo "### ./bin/config/.config.yaml already exists, not overwriting. ###"; \
		echo "##################################################################"; \
	fi

.PHONY: morse_e2e_config_warning
morse_e2e_config_warning: ## Prints a warning if the shannon E2E config is not populated
	@if [ ! -f ./e2e/.morse.config.yaml ]; then \
		echo "#####################################################################"; \
		echo "### Morse E2E config not found, run: 'make copy_morse_e2e_config' ###"; \
		echo "#####################################################################"; \
		exit; \
	fi

# If you are a Grove employee, search for this UUID in 1Password: un76qmz6xunx43icttjmagzlri
.PHONY: copy_morse_e2e_config
copy_morse_e2e_config: ## copies the example Morse test configuration yaml file to .gitignored ..morse.config.yaml file.
	@if [ ! -f ./e2e/.morse.config.yaml ]; then \
		cp ./config/examples/config.morse_example.yaml ./e2e/.morse.config.yaml; \
		echo "#############################################################################################"; \
		echo "### Created ./e2e/.morse.config.yaml                                                      ###"; \
		echo "###                                                                                       ###"; \
		echo "### If you are a Grove employee, use the instructions from the link below to get the      ###"; \
		echo "### correct config file, and COPY IT OVER the ./e2e/.morse.config.yaml file.              ###"; \
		echo "### https://path.grove.city/develop/path/introduction#running-the-e2e-tests-against-morse ###"; \
		echo "###                                                                                       ###"; \
		echo "### Otherwise, please update the the following in .morse.config.yaml:                     ###"; \
		echo "### 'url', 'relay_signing_key', & 'signed_aats'.                                          ###"; \
		echo "#############################################################################################"; \
	else \
		echo "#################################################################"; \
		echo "### ./e2e/.morse.config.yaml already exists, not overwriting. ###"; \
		echo "#################################################################"; \
	fi

.PHONY: config_morse_localnet
config_morse_localnet: ## Create a localnet config file to serve as a Morse gateway
	@if [ ! -f ./local/path/config/.config.yaml ]; then \
		mkdir -p local/path/config; \
		cp ./config/examples/config.morse_example.yaml  local/path/config/.config.yaml; \
		echo "######################################################################"; \
		echo "### Created ./local/path/config/.config.yaml for Morse localnet.   ###"; \
		echo "### README: Please update the the following in .morse.config.yaml: ###"; \
		echo "### 'url', 'relay_signing_key', & 'signed_aats'.                   ###"; \
		echo "######################################################################"; \
	else \
		echo "#########################################################################"; \
		echo "### ./local/path/config/.config.yaml already exists, not overwriting. ###"; \
		echo "#########################################################################"; \
	fi

#########################################
### Envoy Initialization Make Targets ###
#########################################

.PHONY: init_envoy
init_envoy: copy_envoy_config copy_gateway_endpoints ## Runs copy_envoy_config and copy_gateway_endpoints

.PHONY: copy_envoy_config
copy_envoy_config: ## Substitutes the sensitive 0Auth environment variables in the template envoy configuration yaml file and outputs the result to .envoy.yaml
	@mkdir -p local/path/envoy;
	@./envoy/scripts/copy_envoy_config.sh;

.PHONY: copy_gateway_endpoints
copy_gateway_endpoints: ## Copies the example gateway endpoints YAML file from the PADS repo to ./local/path/envoy/.gateway-endpoints.yaml
	@mkdir -p local/path/envoy;
	@./envoy/scripts/copy_gateway_endpoints_yaml.sh;

###############################
### Generation Make Targets ###
###############################

.PHONY: gen_proto
gen_proto: ## Generate the Go code from the gateway_endpoint.proto file
	protoc --go_out=./envoy/auth_server/proto --go-grpc_out=./envoy/auth_server/proto envoy/auth_server/proto/gateway_endpoint.proto

# TODO_IMPROVE(@commoddity): update to use go:generate comments in the interface files and update this target

########################
#### Documentation  ####
########################

.PHONY: go_docs
go_docs: ## Start Go documentation server
	@echo "Visit http://localhost:6060/pkg/github.com/buildwithgrove/path"
	godoc -http=:6060

.PHONY: docusaurus_start
docusaurus_start: ## Start docusaurus server
	cd docusaurus && npm i && npm run start

###############################
###    Makefile imports     ###
###############################

include ./makefiles/localnet.mk
include ./makefiles/quickstart.mk
include ./makefiles/test_requests.mk

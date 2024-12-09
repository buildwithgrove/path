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
####  Build Path Targets  ###
#############################

.PHONY: path_build
path_build: ## build the path binary
	go build -o bin/path ./cmd

.PHONY: path_run
path_run: path_build
	@if [ ! -f ./bin/config/.config.yaml ]; then \
		echo "#########################################################################################"; \
		echo "### ./bin/config/.config.yaml does not exist, use ONE the following to initialize it: ###"; \
		echo "### A. make copy_shannon_config                                                       ###"; \
		echo "### B. make copy_morse_config                                                         ###"; \
		echo "#########################################################################################"; \
		exit; \
	fi; \
	(cd bin; ./path -config ./config/.config.yaml)

.PHONY: path_up
path_up: config_shannon_localnet dev_up config_path_secrets ## Brings up local Tilt development environment (using kind cluster) then runs the PATH gateway and all related dependencies
	tilt up

.PHONY: path_down
path_down: dev_down ## Tears down local Tilt development environment (using kind cluster) then stops the PATH gateway and all related dependencies
	tilt down

#########################
### Test Make Targets ###
#########################

.PHONY: test_all ## Run all tests
test_all: test_unit test_auth_plugin test_e2e_shannon_relay test_e2e_morse_relay

.PHONY: test_unit
test_unit: ## Run all unit tests
	go test ./... -short -count=1

.PHONY: test_auth_server
test_auth_server: ## Run the auth server tests
	(cd envoy/auth_server && go test ./... -count=1)

.PHONY: test_e2e_shannon_relay
test_e2e_shannon_relay: ## Run an E2E shannon relay test
	go test ./... -tags=e2e -count=1 -run Test_ShannonRelay

.PHONY: test_e2e_morse_relay
test_e2e_morse_relay: ## Run an E2E Morse relay test
	go test ./... -tags=e2e -count=1 -run Test_MorseRelay

################################
### Copy Config Make Targets ###
################################

.PHONY: copy_shannon_config
copy_shannon_config: ## copies the example shannon configuration yaml file to .config.yaml file
	@if [ ! -f ./bin/config/.config.yaml ]; then \
		mkdir -p bin/config; \
		cp ./config/examples/config.shannon_example.yaml ./bin/config/.config.yaml; \
		echo "###########################################################################################################################"; \
		echo "### Created ./bin/config/.config.yaml                                                                                   ###"; \
		echo "### README: Please update the the following in .config.yaml: 'gateway_private_key_hex' & 'owned_apps_private_keys_hex'. ###"; \
		echo "###########################################################################################################################"; \
	else \
		echo "##################################################################"; \
		echo "### ./bin/config/.config.yaml already exists, not overwriting. ###"; \
		echo "##################################################################"; \
	fi

.PHONY: copy_morse_config
copy_morse_config: ## copies the example morse configuration yaml file to .config.yaml file
	@if [ ! -f ./bin/config/.config.yaml ]; then \
		mkdir -p bin/config; \
		cp ./config/examples/config.morse_example.yaml ./bin/config/.config.yaml; \
		echo "#############################################################################################################"; \
		echo "### Created ./bin/config/.config.yaml                                                                     ###"; \
		echo "### README: Please update the the following in .config.yaml: 'url', 'relay_signing_key', & 'signed_aats'. ###"; \
		echo "#############################################################################################################"; \
	else \
		echo "##################################################################"; \
		echo "### ./bin/config/.config.yaml already exists, not overwriting. ###"; \
		echo "##################################################################"; \
	fi

.PHONY: copy_shannon_e2e_config
copy_shannon_e2e_config: ## copies the example Shannon test configuration yaml file to .gitignored .shannon.config.yaml file
	@if [ ! -f ./e2e/.shannon.config.yaml ]; then \
		cp ./config/examples/config.shannon_example.yaml ./e2e/.shannon.config.yaml; \
		echo "###################################################################################################################################"; \
		echo "### Created ./e2e/.shannon.config.yaml                                                                                          ###"; \
		echo "### README: Please update the the following in .shannon.config.yaml: 'gateway_private_key_hex' & 'owned_apps_private_keys_hex'. ###"; \
		echo "###################################################################################################################################"; \
	else \
		echo "###################################################################"; \
		echo "### ./e2e/.shannon.config.yaml already exists, not overwriting. ###"; \
		echo "###################################################################"; \
	fi

.PHONY: copy_morse_e2e_config
copy_morse_e2e_config: ## copies the example Morse test configuration yaml file to .gitignored ..morse.config.yaml file.
	@if [ ! -f ./e2e/.morse.config.yaml ]; then \
		cp ./config/examples/config.morse_example.yaml ./e2e/.morse.config.yaml; \
		echo "###################################################################################################################"; \
		echo "### Created ./e2e/.morse.config.yaml                                                                            ###"; \
		echo "### README: Please update the the following in .morse.config.yaml: 'url', 'relay_signing_key', & 'signed_aats'. ###"; \
		echo "###################################################################################################################"; \
	else \
		echo "#################################################################"; \
		echo "### ./e2e/.morse.config.yaml already exists, not overwriting. ###"; \
		echo "#################################################################"; \
	fi

.PHONY: config_shannon_localnet
config_shannon_localnet: ## Create a localnet config file to serve as a Shannon gateway
	@if [ ! -f ./local/path/config/.config.yaml ]; then \
		mkdir -p local/path/config; \
		cp ./config/examples/config.shannon_example.yaml  local/path/config/.config.yaml; \
		echo "###########################################################################################################################"; \
		echo "### Created ./local/path/config/.config.yaml for Shannon localnet.                                         			  ###"; \
		echo "### README: Please update the the following in .config.yaml: 'gateway_private_key_hex' & 'owned_apps_private_keys_hex'. ###"; \
		echo "###########################################################################################################################"; \
	else \
		echo "#########################################################################"; \
		echo "### ./local/path/config/.config.yaml already exists, not overwriting. ###"; \
		echo "#########################################################################"; \
	fi

.PHONY: config_morse_localnet
config_morse_localnet: ## Create a localnet config file to serve as a Morse gateway
	@if [ ! -f ./local/path/config/.config.yaml ]; then \
		mkdir -p local/path/config; \
		cp ./config/examples/config.morse_example.yaml  local/path/config/.config.yaml; \
		echo "######################################################################################################################"; \
		echo "### Created ./local/path/config/.config.yaml for Morse localnet. 										             ###"; \
		echo "### README: Please update the the following in .config.yaml: 'full_node_config.relay_signing_key' & 'signed_aats'. ###"; \
		echo "######################################################################################################################"; \
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
	@if [ ! -f ./local/path/envoy/envoy.yaml ]; then \
		mkdir -p local/path/envoy; \
		./envoy/scripts/copy_envoy_config.sh; \
		echo "###########################################################"; \
		echo "### Created ./local/path/envoy/envoy.yaml               ###"; \
		echo "### README: Please ensure the configuration is correct. ###"; \
		echo "###########################################################"; \
	else \
		echo "######################################################################"; \
		echo "### ./local/path/envoy/envoy.yaml already exists, not overwriting. ###"; \
		echo "######################################################################"; \
	fi

.PHONY: copy_gateway_endpoints
copy_gateway_endpoints: ## Copies the example gateway endpoints YAML file from the PADS repo to ./local/path/envoy/.gateway-endpoints.yaml
	@if [ ! -f ./local/path/envoy/gateway-endpoints.yaml ]; then \
		mkdir -p local/path/envoy; \
		./envoy/scripts/copy_gateway_endpoints_yaml.sh; \
		echo "###########################################################"; \
		echo "### Created ./local/path/envoy/gateway-endpoints.yaml   ###"; \
		echo "### README: Please update this file with your own data. ###"; \
		echo "###########################################################"; \
	else \
		echo "##################################################################################"; \
		echo "### ./local/path/envoy/gateway-endpoints.yaml already exists, not overwriting. ###"; \
		echo "##################################################################################"; \
	fi

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

.PHONY: docs_update
## TODO_UPNEXT(@HebertCL): handle documentation update like poktroll
docs_update: ## Update documentation from README.
	cat README.md > docusaurus/docs/README.md

.PHONY: docusaurus_start
docusaurus_start: ## Start docusaurus server
	cd docusaurus && npm i && npm run start

###############################
###    Makefile imports     ###
###############################

include ./makefiles/localnet.mk

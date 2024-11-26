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
### Run Path Make Targets ###
#############################

.PHONY: path_up_gateway
path_up_gateway: ## Run just the PATH gateway without any dependencies
	docker compose --profile path-gateway up -d --no-deps path_gateway 

.PHONY: path_up_build_gateway
path_up_build_gateway: ## Run and build just the PATH gateway without any dependencies
	docker compose --profile path-gateway up -d --build --no-deps path_gateway

.PHONY: path_down_gateway
path_down_gateway: ## Stop just the PATH gateway
	docker compose --profile path-gateway down --remove-orphans path_gateway

.PHONY: path_up
path_up: ## Run the PATH gateway and all related dependencies
	docker compose up -d

.PHONY: path_up_build
path_up_build: ## Run and build the PATH gateway and all related dependencies
	docker compose up -d --build

.PHONY: path_down
path_down: ## Stop the PATH gateway and all related dependencies
	docker compose down --remove-orphans

#########################
### Test Make Targets ###
#########################

.PHONY: test_all ## Run all tests
test_all: test_unit test_auth_plugin test_e2e_shannon_relay

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
	@if [ ! -f ./cmd/.config.yaml ]; then \
		cp ./cmd/.config.shannon_example.yaml ./cmd/.config.yaml; \
	else \
		echo ".config.yaml already exists, not overwriting."; \
	fi

.PHONY: copy_morse_config
copy_morse_config: ## copies the example morse configuration yaml file to .config.yaml file
	@if [ ! -f ./cmd/.config.yaml ]; then \
		cp ./cmd/.config.morse_example.yaml ./cmd/.config.yaml; \
	else \
		echo ".config.yaml already exists, not overwriting."; \
	fi

.PHONY: copy_shannon_e2e_config
copy_shannon_e2e_config: ## copies the example Shannon test configuration yaml file to .gitignored .shannon.config.yaml file
	@if [ ! -f ./e2e/.shannon.config.yaml ]; then \
		cp ./e2e/shannon.example.yaml ./e2e/.shannon.config.yaml; \
	else \
		echo "./e2e/.shannon.config.yaml already exists, not overwriting."; \
	fi

.PHONY: copy_morse_e2e_config
copy_morse_e2e_config: ## copies the example Morse test configuration yaml file to .gitignored ..morse.config.yaml file.
	@if [ ! -f ./e2e/.morse.config.yaml ]; then \
		cp ./e2e/morse.example.yaml ./e2e/.morse.config.yaml; \
	else \
		echo "./e2e/.morse.config.yaml already exists, not overwriting."; \
	fi

.PHONY: copy_envoy_config
copy_envoy_config: ## substitutes the sensitive Auth0 environment variables in the template envoy configuration yaml file and outputs the result to .envoy.yaml
	@if [ ! -f ./envoy/envoy.yaml ]; then \
		./envoy/scripts/copy_envoy_config.sh; \
	else \
		echo "./envoy/envoy.yaml already exists, not overwriting."; \
	fi

.PHONY: copy_envoy_env
copy_envoy_env: ## copies the example envoy environment variables file to .env file
	@if [ ! -f ./envoy/auth_server/.env ]; then \
		cp ./envoy/auth_server/.env.example ./envoy/auth_server/.env; \
	else \
		echo "./envoy/auth_server/.env already exists, not overwriting."; \
	fi

.PHONY: copy_envoy_gateway_endpoints
copy_envoy_gateway_endpoints: ## copies the example envoy gateway endpoints file to gateway-endpoints.yaml
	@if [ ! -f ./envoy/gateway-endpoints.yaml ]; then \
		cp ./envoy/gateway-endpoints.example.yaml ./envoy/gateway-endpoints.yaml; \
	else \
		echo "./envoy/gateway-endpoints.yaml already exists, not overwriting."; \
	fi

.PHONY: init_envoy
init_envoy: ## copies the example envoy configuration and gateway endpoints files to their respective files
	@make copy_envoy_config
	@make copy_envoy_env
	@make copy_envoy_gateway_endpoints

###############################
### Generation Make Targets ###
###############################

.PHONY: gen_proto
gen_proto: ## Generate the Go code from the gateway_endpoint.proto file
	protoc --go_out=./envoy/auth_server/proto --go-grpc_out=./envoy/auth_server/proto envoy/auth_server/proto/gateway_endpoint.proto

# TODO_IMPROVE(@commoddity): update to use go:generate comments in the interface files and update this target
# TODO_TECHDEBT(@commoddity): move all mocks to a shared mocks package
.PHONY: gen_mocks
gen_mocks: ## Generate the mock code from the gateway_endpoint.proto file
	mockgen -source=./envoy/auth_server/proto/gateway_endpoint_grpc.pb.go -destination=./envoy/auth_server/endpoint_store/client_mock_test.go -package=endpointstore -mock_names=GatewayEndpointsClient=MockGatewayEndpointsClient

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

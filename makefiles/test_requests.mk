#######################
#### Test Requests ####
#######################

# This Makefile provides examples of the various ways to make requests to PATH.

# NOTE: All of these requests assume a Shannon Gateway, as the service ID is 'anvil'.

.PHONY: debug_relayminer_supplier_info_msg
debug_relayminer_supplier_info_msg: ## Displays debugging guidance for Anvil supplier issues
	@echo "#######################################################################################################################################"
	@echo "INFO: If a request did not succeed, look into debugging the Anvil supplier by reviewing:"
	@echo "  https://www.notion.so/buildwithgrove/PATH-Shannon-Beta-Critical-Relay-Miner-Infrastructure-for-PATH-Supplier-Anvil-E2E-17da36edfff680da98f2ff01705be00b"
	@echo "########################################################################################################################################"

.PHONY: check_path_up_with_envoy
check_path_up_with_envoy: ## Checks if PATH with Envoy is running at localhost:3070
	@if ! nc -z localhost 3070 2>/dev/null; then \
		@echo "########################################################################"; \
		@echo "ERROR: Envoy PATH is not running on port 3070"; \
		@echo "Please start it with:"; \
		@echo "  make path_up"; \
		@echo "########################################################################"; \
		exit 1; \
	fi

.PHONY: check_path_up_without_envoy
check_path_up_without_envoy: ## Checks if standalone PATH (without GUARD) is running at localhost:3069
	@if ! nc -z localhost 3069 2>/dev/null; then \
		@echo "########################################################################"; \
		@echo "ERROR: Standalone PATH is not currently running on localhost:3069"; \
		@echo "Please start it with:"; \
		@echo "  make path_run"; \
		@echo "########################################################################"; \
		exit 1; \
	fi


####################################
#### PATH + GUARD Test Requests ####
####################################

# For all of the below requests:
# - The full PATH stack including GUARD must be running
# - The 'anvil' service must be configured in the '.config.yaml' file.

# The following are the various ways to make requests to PATH with Envoy running:
# - **Auth**: static API key, passed in the 'Authorization' header
# - **Service ID**: passed as the subdomain or in the 'Target-Service-Id' header

.PHONY: test_request__service_id_subdomain
test_request__service_id_subdomain: check_path_up_with_envoy debug_anvil_supplier_info_msg ## Test request with API key auth and the service ID passed as the subdomain
	curl http://anvil.localhost:3070/v1 \
		-H "Authorization: test_api_key" \
		-d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'

.PHONY: test_request__service_id_header_shannon
test_request__service_id_header_shannon: check_path_up_with_envoy debug_anvil_supplier_info_msg ## Test request with API key auth and the service ID passed in the Target-Service-Id header
	curl http://localhost:3070/v1 \
		-H "Target-Service-Id: anvil" \
		-H "Authorization: test_api_key" \
		-d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'

.PHONY: test_request__service_id_header_morse
test_request__service_id_header_morse: check_path_up_with_envoy debug_anvil_supplier_info_msg ## Test request with API key auth and the service ID passed in the Target-Service-Id header
	curl http://localhost:3070/v1 \
		-H "Target-Service-Id: polygon" \
		-H "Authorization: test_api_key" \
		-d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'

############################
#### PATH Test Requests ####
############################

.PHONY: test_request__evm_endpoint
test_request__evm_endpoint: check_path_up_without_envoy debug_relayminer_supplier_info_msg ## Test EVM endpoint request against the PATH Gateway running on port 3069 without GUARD
	curl http://localhost:3069/v1/ \
		-H "Target-Service-Id: anvil" \
		-d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'

.PHONY: test_request__cometbft_endpoint
test_request__cometbft_endpoint: check_path_up_without_envoy ## Test CometBFT endpoint request against the PATH Gateway running on port 3069 without GUARD
	curl 'http://localhost:3069/v1/status' \
		-H 'Target-Service-Id: cometbft'

###################################
#### Relay Utils Test Requests ####
###################################

.PHONY: check_relay_util
# Internal helper: Checks if relay-util is installed locally
check_relay_util:
	@if ! command -v relay-util &> /dev/null; then \
		echo "####################################################################################################"; \
		echo "Relay Util is not installed." \
		echo "To use any Relay Util make targets to send load testing requests please install Relay Util with:"; \
		echo "go install github.com/commoddity/relay-util/v2@latest"; \
		echo "####################################################################################################"; \
	fi

.PHONY: test_request__relay_util_10
test_request__relay_util_10: check_path_up_without_envoy check_relay_util ## Test anvil via PATH with 10 eth_blockNumber requests using relay-util
	relay-util \
		-u http://localhost:3069/v1 \
		-H "target-service-id: anvil" \
		-d '{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}' \
		-x 10 \
		-b

.PHONY: test_request__relay_util_1000
test_request__relay_util_1000: check_path_up_without_envoy check_relay_util  ## Test anvil via PATH with 10,000 eth_blockNumber requests using relay-util
	relay-util \
		-u http://localhost:3069/v1 \
		-H "target-service-id: anvil" \
		-d '{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}' \
		-x 1000 \
		-b

.PHONY: test_request__relay_util_10_via_guard
test_request__relay_util_10_via_guard: check_path_up_with_envoy check_relay_util  ## Test anvil PATH behind GUARD with 10 eth_blockNumber requests using relay-util
	relay-util \
		-u http://localhost:3070/v1 \
		-H "target-service-id: anvil" \
		-H "authorization: test_api_key" \
		-d '{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}' \
		-x 10 \
		-b

.PHONY: test_request__relay_util_1000_via_envoy
test_request__relay_util_1000_via_envoy: check_path_up_with_envoy check_relay_util  ## Test anvil via PATH behind GUARD with 10,000 eth_blockNumber requests using relay-util
	relay-util \
		-u http://localhost:3070/v1 \
		-H "target-service-id: anvil" \
		-H "authorization: test_api_key" \
		-d '{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}' \
		-x 1000 \
		-b


.PHONY: test_request__relay_util_100_F00C
test_request__relay_util_100_F00C: check_path_up_without_envoy check_relay_util  ## Test F00C (Eth MainNet on Morse) via PATH with 10,000 eth_blockNumber requests using relay-util
	relay-util \
		-u http://localhost:3069/v1 \
		-H "target-service-id: F00C" \
		-H "authorization: test_api_key" \
		-d '{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}' \
		-x 100 \
		-b


.PHONY: test_request__relay_util_100_F00C_via_envoy
test_request__relay_util_100_F00C_via_envoy: check_path_up_with_envoy check_relay_util  ## Test F00C (Eth MainNet on Morse) via PATH behind GUARD with 10,000 eth_blockNumber requests using relay-util
	relay-util \
		-u http://localhost:3070/v1 \
		-H "target-service-id: F00C" \
		-H "authorization: test_api_key" \
		-d '{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}' \
		-x 100 \
		-b
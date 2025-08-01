# This Makefile provides examples of the various ways to make requests to PATH.

#################
#### Helpers ####
#################

.PHONY: debug_relayminer_supplier_info_msg
# Internal helper: Displays debugging guidance for Anvil supplier issues
debug_relayminer_supplier_info_msg:
	@echo "#######################################################################################################################################"
	@echo "INFO: If a request did not succeed, look into debugging the Anvil supplier by reviewing:"
	@echo "https://www.notion.so/buildwithgrove/PATH-Shannon-Beta-Critical-Relay-Miner-Infrastructure-for-PATH-Supplier-Anvil-E2E-17da36edfff680da98f2ff01705be00b"
	@echo "########################################################################################################################################"
	@echo ""

.PHONY: debug_view_results_links
# Internal helper: Displays links to view results in local dashboards
debug_view_results_links:
	@echo "##########################################################################################################"
	@echo "####   VIEW RESULTS IN LOCAL DASHBOARDS   ####"
	@echo ""
	@echo "1. Path Service Requests:"
	@echo "   http://localhost:3003/d/relays/path-service-requests?orgId=1&from=now-15m&to=now&timezone=browser"
	@echo ""
	@echo "2. Path Gateway Metrics:"
	@echo "   http://localhost:3003/d/gateway/path-path-gateway?orgId=1&from=now-1h&to=now&timezone=browser&var-path=path-metrics&refresh=5s"
	@echo ""
	@echo "3. Morse Relay Requests:"
	@echo "   http://localhost:3003/d/morse/morse-relay-requests?orgId=1&from=now-3h&to=now&timezone=browser&refresh=10s"
	@echo ""
	@echo "Login with: admin / admin (for now)"
	@echo "##########################################################################################################"
	@echo ""

.PHONY: check_path_up_binary
# Internal helper: Checks if PATH is running at localhost:3069
check_path_up_binary:
	@if ! nc -z localhost 3069 2>/dev/null; then \
		echo "########################################################################"; \
		echo "ERROR: PATH is not running on port 3069"; \
		echo "Please start it with:"; \
		echo "  make path_run"; \
		echo "########################################################################"; \
		exit 1; \
	fi

.PHONY: check_path_up_envoy
# Internal helper: Checks if PATH + GUARD is running at localhost:3070
check_path_up_envoy:
	@if ! nc -z localhost 3070 2>/dev/null; then \
		echo "########################################################################"; \
		echo "ERROR: PATH is not running on port 3070"; \
		echo "Please start it with:"; \
		echo "  make path_up"; \
		echo "########################################################################"; \
		exit 1; \
	fi

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

###################################
#### PATH binary Test Requests ####
###################################

# For all of the below requests:
# - The PATH binary must be running
#   - Run the PATH binary with:
#     `make path_run``
# - The PATH binary will be available at `localhost:3069`

# For all of the below requests:
# - The 'eth' service must be configured in the '.config.yaml' file.
# - The application must be configured to serve requests for `eth` (Eth MainNet on Shannon)
# - It is assumed that the network has suppliers running that service `eth` requests

# The following are the various ways to make requests to PATH with the PATH binary running:
# - Service ID: passed as the subdomain or in the 'Target-Service-Id' header

.PHONY: test_healthz__binary
test_healthz__binary: check_path_up_binary ## Test healthz request to PATH binary
	curl http://localhost:3069/healthz

.PHONY: test_request__binary__eth
test_request__binary__eth: check_path_up_binary ## Test single eth_blockNumber request to PATH binary with service ID in header
	curl http://localhost:3069/v1 \
		-H "Target-Service-Id: eth" \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber"}'

.PHONY: test_request__binary__eth__batch
test_request__binary__eth__batch: check_path_up_binary ## Test batch request (eth_blockNumber, eth_chainId, eth_gasPrice) to PATH binary
	curl http://localhost:3069/v1 \
		-H "Target-Service-Id: eth" \
		-H "Content-Type: application/json" \
		-d '[{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber"}, {"jsonrpc": "2.0", "id": 2, "method": "eth_chainId"}, {"jsonrpc": "2.0", "id": 3, "method": "eth_gasPrice"}]'

.PHONY: test_request__binary__relay_util__eth
test_request__binary__relay_util__eth: check_path_up_binary check_relay_util  ## Test eth PATH binary with 100 eth_blockNumber requests using relay-util. Override service by running: SERVICE_ID=eth make test_request__binary__relay_util__eth
	relay-util \
		-u http://localhost:3069/v1 \
		-H "Target-Service-Id: $${SERVICE_ID:-eth}" \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}' \
		-x 100 \
		-b

.PHONY: test_disqualified_endpoints__binary
test_disqualified_endpoints__binary: check_path_up_binary ## Get list of currently disqualified eth endpoints with reasons
	curl http://localhost:3069/disqualified_endpoints \
		-H "Target-Service-Id: eth"

####################################
#### PATH + GUARD Test Requests ####
####################################

# For all of the below requests:
# - The full PATH stack (including GUARD) must be running
#   - Run the full PATH stack with:
#     `make path_up`
# - The full PATH stack will be available at `localhost:3070`

# For all of the below requests:
# - The 'eth' service must be configured in the '.config.yaml' file.
# - The application must be configured to serve requests for `eth` (Eth MainNet on Shannon)
# - It is assumed that the network has suppliers running that service `eth` requests

# The following are the various ways to make requests to PATH with Envoy running:
# - Auth: static API key, passed in the 'Authorization' header
# - Service ID: passed as the subdomain or in the 'Target-Service-Id' header

.PHONY: test_healthz__envoy
test_healthz__envoy: check_path_up_envoy ## Test healthz request to PATH + GUARD
	curl http://localhost:3070/healthz

.PHONY: test_request__envoy_subdomain__eth
test_request__envoy_subdomain__eth: check_path_up_envoy debug_relayminer_supplier_info_msg ## Test request with API key auth and the service ID passed as the subdomain
	curl http://eth.localhost:3070/v1 \
		-H "Authorization: test_api_key" \
		-d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'

.PHONY: test_request__envoy_header__eth
test_request__envoy_header__eth: check_path_up_envoy debug_relayminer_supplier_info_msg ## Test request with API key auth and the service ID passed in the Target-Service-Id header
	curl http://localhost:3070/v1 \
		-H "Target-Service-Id: eth" \
		-H "Authorization: test_api_key" \
		-d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber"}'

.PHONY: test_request__envoy_subdomain__eth_batch
test_request__envoy_subdomain__eth_batch: check_path_up_envoy debug_relayminer_supplier_info_msg ## Test batch request with API key auth and service ID as subdomain
	curl http://eth.localhost:3070/v1 \
		-H "Authorization: test_api_key" \
		-H "Content-Type: application/json" \
		-d '[{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber"}, {"jsonrpc": "2.0", "id": 2, "method": "eth_chainId"}, {"jsonrpc": "2.0", "id": 3, "method": "eth_gasPrice"}]'

.PHONY: test_request__envoy_header__eth_batch
test_request__envoy_header__eth_batch: check_path_up_envoy debug_relayminer_supplier_info_msg ## Test batch request with API key auth and service ID in header
	curl http://localhost:3070/v1 \
		-H "Target-Service-Id: eth" \
		-H "Authorization: test_api_key" \
		-H "Content-Type: application/json" \
		-d '[{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber"}, {"jsonrpc": "2.0", "id": 2, "method": "eth_chainId"}, {"jsonrpc": "2.0", "id": 3, "method": "eth_gasPrice"}]'

.PHONY: test_request__envoy_relay_util__eth
test_request__envoy_relay_util__eth: check_path_up_envoy check_relay_util debug_view_results_links  ## Test eth PATH behind GUARD with 100 eth_blockNumber requests using relay-util. Override service by running: SERVICE_ID=eth make test_request__envoy_relay_util__eth
	relay-util \
		-u http://localhost:3070/v1 \
		-H "Target-Service-Id: $${SERVICE_ID:-eth}" \
		-H "Authorization: test_api_key" \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}' \
		-x 100 \
		-b

.PHONT: test_disqualified_endpoints__envoy
test_disqualified_endpoints__envoy: check_path_up_envoy ## Get list of currently disqualified eth endpoints with reasons
	curl http://localhost:3070/disqualified_endpoints \
		-H "Target-Service-Id: eth"

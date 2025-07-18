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

.PHONY: check_path_up
# Internal helper: Checks if PATH is running at localhost:3070
check_path_up:
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

####################################
#### PATH + GUARD Test Requests ####
####################################

# For all of the below requests:
# - The full PATH stack (including GUARD) must be running

# For all Shannon requests:
# - The 'anvil' service must be configured in the '.config.yaml' file.
# - The application must be configured to serve requests for `anvil` (Eth MainNet on Shannon)
# - It is assumed that the network has suppliers running that service `anvil` requests

# For all Morse requests:
# - The application must be configured to serve requests for `F00C` (Eth MainNet on Morse)
# - It is assumed that the network has suppliers running that service `F00C` requests

# The following are the various ways to make requests to PATH with Envoy running:
# - Auth: static API key, passed in the 'Authorization' header
# - Service ID: passed as the subdomain or in the 'Target-Service-Id' header

.PHONY: test_request__shannon_service_id_subdomain
test_request__shannon_service_id_subdomain: check_path_up debug_relayminer_supplier_info_msg ## Test request with API key auth and the service ID passed as the subdomain
	curl http://anvil.localhost:3070/v1 \
		-H "Authorization: test_api_key" \
		-d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'

.PHONY: test_request__shannon_service_id_header
test_request__shannon_service_id_header: check_path_up debug_relayminer_supplier_info_msg ## Test request with API key auth and the service ID passed in the Target-Service-Id header
	curl http://localhost:3070/v1 \
		-H "Target-Service-Id: anvil" \
		-H "Authorization: test_api_key" \
		-d '{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }'

##################################
#### Relay Util Test Requests ####
##################################

.PHONY: test_request__shannon_relay_util_100
test_request__shannon_relay_util_100: check_path_up check_relay_util debug_view_results_links  ## Test anvil PATH behind GUARD with 10 eth_blockNumber requests using relay-util. Override service by running: SERVICE_ID=eth make test_request__shannon_relay_util_100
	relay-util \
		-u http://localhost:3070/v1 \
		-H "target-service-id: $${SERVICE_ID:-anvil}" \
		-H "authorization: test_api_key" \
		-d '{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}' \
		-x 100 \
		-b
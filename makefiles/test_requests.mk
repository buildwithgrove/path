# This Makefile provides examples of the various ways to make requests to PATH.

#################
#### Helpers ####
#################

.PHONY: debug_relayminer_supplier_info_msg
# Internal helper: Displays debugging guidance for Anvil supplier issues
debug_relayminer_supplier_info_msg:
	@echo "#######################################################################################################################################"
	@echo "INFO: If a request did not succeed, look into debugging the Anvil supplier by reviewing:"
	@echo "${CYAN}https://www.notion.so/buildwithgrove/PATH-Shannon-Beta-Critical-Relay-Miner-Infrastructure-for-PATH-Supplier-Anvil-E2E-17da36edfff680da98f2ff01705be00b${RESET}"
	@echo "########################################################################################################################################"
	@echo ""

.PHONY: debug_view_results_links
# Internal helper: Displays links to view results in local dashboards
debug_view_results_links:
	@echo "##############################################"
	@echo "####   ${BLUE}VIEW RESULTS IN LOCAL DASHBOARDS${RESET}   ####"
	@echo "##############################################"
	@echo ""
	@echo "1. Path Service Requests: ${CYAN}http://localhost:3003/d/relays/path-service-requests?orgId=1&from=now-15m&to=now&timezone=browser${RESET}"
	@echo ""
	@echo "2. Path Gateway Metrics: ${CYAN}http://localhost:3003/d/gateway/path-path-gateway?orgId=1&from=now-1h&to=now&timezone=browser&var-path=path-metrics&refresh=5s${RESET}"
	@echo ""
	@echo "${BOLD}Login with: admin / admin (for now)${RESET}"
	@echo "##############################################"
	@echo ""

####################################
#### PATH + GUARD Test Requests ####
####################################

# For all of the below requests:
# - The full PATH stack (including GUARD) must be running

# For all requests:
# - The 'anvil' service must be configured in the '.config.yaml' file.
# - The application must be configured to serve requests for `anvil` (Eth MainNet on Shannon)
# - It is assumed that the network has suppliers running that service `anvil` requests

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

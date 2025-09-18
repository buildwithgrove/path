# Examples of running load tests against PATH & Portal

.PHONY: check_websocket_load_test
# Internal helper: Checks if websocket-load-test is installed locally
check_websocket_load_test:
	@if ! command -v websocket-load-test &> /dev/null; then \
		echo "####################################################################################################"; \
		echo "Websocket Load Test is not installed." \
		echo "To use any Websocket Load Test make targets to send load testing requests please install Websocket Load Test with:"; \
		echo "go install github.com/commoddity/websocket-load-test@latest"; \
		echo "####################################################################################################"; \
	fi

.PHONY: test_load__relay_util__local
test_load__relay_util__local: check_path_up check_relay_util debug_view_results_links  ## Load test an anvil endpoint with PATH behind GUARD with 10 eth_blockNumber requests. Override service by running: SERVICE_ID=eth make test_load__relay_util__local
	relay-util \
		-u http://localhost:3070/v1 \
		-H "target-service-id: $${SERVICE_ID:-eth}" \
		$${GROVE_PORTAL_APP_ID:+-H "Portal-Application-ID: $${GROVE_PORTAL_APP_ID}"} \
		-H "authorization: $${GROVE_PORTAL_API_KEY:-test_api_key}" \
		-d '{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}' \
		-x 100 \
		-b

.PHONY: test_load__websocket_load_test__local
test_load__websocket_load_test__local: check_path_up check_websocket_load_test debug_view_results_links  ## Load test a websocket connection with subscriptions to newHeads and newPendingTransactions.
	websocket-load-test \
	   --service "$${SERVICE_ID:-xrplevm}" \
	   --app-id $$GROVE_PORTAL_APP_ID \
	   --api-key $$GROVE_PORTAL_API_KEY \
	   --subs "newHeads,newPendingTransactions" \
	   --count 10 \
	   --log

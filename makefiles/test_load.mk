# Examples of running load tests against PATH & Portal

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

.PHONY: check_websocket_load_test
# Internal helper: Checks if websocket-load-test is installed locally
check_websocket_load_test:
	@if ! command -v websocket-load-test &> /dev/null; then \
		echo "####################################################################################################"; \
		echo "WebSocket Load Test is not installed." \
		echo "To use any WebSocket Load Test make targets to send load testing requests please install WebSocket Load Test with:"; \
		echo "go install github.com/commoddity/websocket-load-test@latest"; \
		echo "####################################################################################################"; \
	fi

.PHONY: test_load__relay_util__local
test_load__relay_util__local: check_path_up check_relay_util debug_view_results_links  ## Load test an anvil endpoint with PATH behind GUARD with 10 eth_blockNumber requests. Override service by running: SERVICE_ID=eth make test_load__relay_util__local
	relay-util \
		-u http://localhost:3070/v1 \
		-H "target-service-id: $${SERVICE_ID:-anvil}" \
		-H "authorization: test_api_key" \
		-d '{"jsonrpc":"2.0","method":"eth_blockNumber","id":1}' \
		-x 100 \
		-b

# TODO_IN_THIS_PR: Finish this.
# .PHONY: test_load__websocket_load_test__local
# test_load__websocket_load_test__local: check_path_up check_websocket_load_test debug_view_results_links  ## Load test a websocket connection.
# websocket-load-test \

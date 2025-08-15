# DEV_NOTE: DO NOT CHANGE the (cd e2e && go test ...) to the (go test ... e2e)
# in the helpers below. This is needed to ensure the logs are beautified as expected.

#########################
### Test Make Targets ###
#########################

.PHONY: test_all ## Run all unit tests and E2E test a subset of key services.
test_all: test_unit
	@$(MAKE) e2e_test eth,poly,xrplevm-testnet,oasys

.PHONY: test_unit
test_unit: ## Run all unit tests
	go test ./... -short -count=1

.PHONY: go_lint
go_lint: ## Run all go linters
	golangci-lint run --timeout 5m --build-tags test

#################
### E2E Tests ###
#################

# HTTP E2E Tests
.PHONY: e2e_test_all
e2e_test_all: shannon_e2e_config_warning ## Run HTTP-only E2E tests for all service IDs
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=shannon go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

.PHONY: e2e_test
e2e_test: shannon_e2e_config_warning ## Run HTTP-only E2E tests with specified service IDs (e.g. make e2e_test eth,xrplevm)
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "❌ Error: Service IDs are required (comma-separated list)"; \
		echo "  👀 Example: make test_e2e_evm_shannon eth,xrplevm"; \
		echo "  💡 To run with default service IDs, use: make test_e2e_evm_shannon_defaults"; \
		exit 1; \
	fi
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=shannon TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

# WebSocket E2E Tests
.PHONY: e2e_test_websocket_all
e2e_test_websocket_all: shannon_e2e_config_warning ## Run WebSocket-only E2E tests for all WebSocket-compatible services
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=shannon TEST_WEBSOCKETS=true go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

.PHONY: e2e_test_websocket
e2e_test_websocket: shannon_e2e_config_warning ## Run WebSocket-only E2E tests with specified service IDs (e.g. make e2e_test_websocket xrplevm)
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "❌ Error: Service IDs are required (comma-separated list)"; \
		echo "  👀 Example: make e2e_test_websocket xrplevm,xrplevm-testnet"; \
		echo "  💡 To run all WebSocket-compatible services, use: make e2e_test_websocket_all"; \
		exit 1; \
	fi
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=shannon TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) TEST_WEBSOCKETS=true go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

.PHONY: e2e_test_eth_fallback
e2e_test_eth_fallback: shannon_e2e_config_warning ## Run an E2E Shannon relay test with ETH fallback enabled (requires FALLBACK_URL)
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "❌ Error: ETH fallback URL is required"; \
		echo "  👀 Example: make e2e_test_eth_fallback https://eth.rpc.backup.io"; \
		echo "  💡 Usage: make e2e_test_eth_fallback <FALLBACK_URL>"; \
		echo "  📝 The fallback URL should be a valid HTTP/HTTPS endpoint for ETH service"; \
		exit 1; \
	fi
	./e2e/scripts/run_eth_fallback_test.sh $(filter-out $@,$(MAKECMDGOALS))

##################
### Load Tests ###
##################

# HTTP Load Tests
.PHONY: load_test_all
load_test_all: ## Run HTTP-only load tests for all service IDs
	(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

.PHONY: load_test
load_test: ## Run HTTP-only load tests with specified service IDs (e.g. make load_test eth,xrplevm)
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "❌ Error: Service IDs are required (comma-separated list)"; \
		echo "  👀 Example: make load_test eth,xrplevm"; \
		echo "  💡 To run with default service IDs, use: make load_test_defaults"; \
		exit 1; \
	fi
	@(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

# WebSocket Load Tests
.PHONY: load_test_websocket_all
load_test_websocket_all: ## Run WebSocket-only load tests for all WebSocket-compatible services
	(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon TEST_WEBSOCKETS=true go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

.PHONY: load_test_websocket
load_test_websocket: ## Run WebSocket-only load tests with specified service IDs (e.g. make load_test_websocket xrplevm)
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "❌ Error: Service IDs are required (comma-separated list)"; \
		echo "  👀 Example: make load_test_websocket xrplevm,xrplevm-testnet"; \
		echo "  💡 To run all WebSocket-compatible services, use: make load_test_websocket_all"; \
		exit 1; \
	fi
	@(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) TEST_WEBSOCKETS=true go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

.PHONY: config_enable_grove_fallback
config_enable_grove_fallback: ## Enable fallback endpoints for all services in PATH config
	@echo "🔧 Enabling fallback endpoints for all services..."
	@echo "📝 Updating local/path/.config.yaml to send all traffic to fallback endpoints"
	@yq eval '.shannon_config.gateway_config.service_fallback[].send_all_traffic = true' -i local/path/.config.yaml
	@echo "✅ Fallback endpoints enabled for all services"

# In order to allow passing the service IDs to the load test targets, this target is needed to avoid printing an error.
%:
	@:
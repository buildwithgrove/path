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

.PHONY: e2e_test_all
e2e_test_all: shannon_e2e_config_warning ## Run an E2E Shannon relay test for all service IDs
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=shannon go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

.PHONY: e2e_test
e2e_test: shannon_e2e_config_warning ## Run an E2E Shannon relay test with specified service IDs (e.g. make shannon_test_e2e eth,xrplevm)
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "❌ Error: Service IDs are required (comma-separated list)"; \
		echo "  👀 Example: make test_e2e_evm_shannon eth,xrplevm"; \
		echo "  💡 To run with default service IDs, use: make test_e2e_evm_shannon_defaults"; \
		exit 1; \
	fi
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=shannon TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

.PHONY: e2e_test_eth_fallback_ci
e2e_test_eth_fallback_ci: shannon_e2e_config_warning ## Run an E2E Shannon relay test with ETH fallback enabled
	@echo "⚠️  WARNING: This target modifies .shannon.config.yaml and should primarily be used in CI environments! 🤖"
	@echo "🔧 This will configure ETH fallback endpoints for testing against external providers"
	@echo "📝 Make sure SHANNON_ETH_FALLBACK_URL environment variable is set"
	@echo "🔄 Updating .shannon.config.yaml with fallback settings..."
	@echo ""
	./e2e/scripts/update_shannon_config_fallback_eth.sh
	@$(MAKE) e2e_test eth

##################
### Load Tests ###
##################

.PHONY: load_test_all
load_test_all: ## Run a Shannon load test for all service IDs
	(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

.PHONY: load_test
load_test: ## Run a Shannon load test with specified service IDs (e.g. make load_test eth,xrplevm)
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "❌ Error: Service IDs are required (comma-separated list)"; \
		echo "  👀 Example: make load_test eth,xrplevm"; \
		echo "  💡 To run with default service IDs, use: make load_test_defaults"; \
		exit 1; \
	fi
	@(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

# In order to allow passing the service IDs to the load test targets, this target is needed to avoid printing an error.
%:
	@:

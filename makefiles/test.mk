# DEV_NOTE: DO NOT CHANGE the (cd e2e && go test ...) to the (go test ... e2e)
# in the helpers below. This is needed to ensure the logs are beautified as expected.

#########################
### Test Make Targets ###
#########################

# TODO_TECHDEBT(@commoddity): Remove Morse test targets after Shannon migration.

.PHONY: test_all ## Run all unit tests and E2E test a subset of key services.
test_all: test_unit
	@$(MAKE) e2e_test eth,poly,xrplevm-testnet,oasys
	@$(MAKE) morse_e2e_test F00C,F021,F036,F01C

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
e2e_test: shannon_e2e_config_warning ## Run an E2E Shannon relay test with specified service IDs (e.g. make shannon_test_e2e eth,anvil)
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "❌ Error: Service IDs are required (comma-separated list)"; \
		echo "  👀 Example: make test_e2e_evm_shannon eth,anvil"; \
		echo "  💡 To run with default service IDs, use: make test_e2e_evm_shannon_defaults"; \
		exit 1; \
	fi
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=shannon TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

.PHONY: morse_e2e_test_all
morse_e2e_test_all: morse_e2e_config_warning ## Run an E2E Morse relay test for all service IDs
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=morse go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

.PHONY: morse_e2e_test
morse_e2e_test: morse_e2e_config_warning ## Run an E2E Morse relay test with specified service IDs (e.g. make morse_test_e2e F00C,F021)
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "❌ Error: Service IDs are required (comma-separated list)"; \
		echo "  👀 Example: make test_e2e_evm_morse F00C,F021"; \
		echo "  💡 To run with default service IDs, use: make test_e2e_evm_morse_defaults"; \
		exit 1; \
	fi
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=morse TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

##################
### Load Tests ###
##################

# Shannon load tests use the simpler `load_test` targets as Shannon is the main focus of the load testing tool.

.PHONY: load_test_all
load_test_all: ## Run a Shannon load test for all service IDs
	(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

.PHONY: load_test
load_test: ## Run a Shannon load test with specified service IDs (e.g. make load_test eth,anvil)
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "❌ Error: Service IDs are required (comma-separated list)"; \
		echo "  👀 Example: make load_test eth,anvil"; \
		echo "  💡 To run with default service IDs, use: make load_test_defaults"; \
		exit 1; \
	fi
	@(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E)


##################
### Portal Tests ###
##################

.PHONY: make_load_test_portal_local_path
make_load_test_portal_local_path: ## Run a Shannon load test against local PATH with Portal credentials
	@echo "📋 Copying archival config..."
	@cp e2e/config/.grove.archival.e2e_load_test.config.yaml e2e/config/.e2e_load_test.config.yaml
	@echo "✅ Config copied successfully"
	@echo ""
	@echo "🔍 Checking environment variables..."
	@if [ -z "$(PORTAL_APPLICATION_ID)" ]; then \
		echo "❌ Error: PORTAL_APPLICATION_ID environment variable is not set"; \
		echo "  🔐 Grove employees: Get credentials from https://share.1password.com/s#x4x-cWnkf4GUw50BwV166aWEEevZV1nmaO2RxJssWjg"; \
		echo "  💡 Set with: export PORTAL_APPLICATION_ID=your_app_id"; \
		exit 1; \
	fi
	@if [ -z "$(PORTAL_API_KEY)" ]; then \
		echo "❌ Error: PORTAL_API_KEY environment variable is not set"; \
		echo "  🔐 Grove employees: Get credentials from https://share.1password.com/s#x4x-cWnkf4GUw50BwV166aWEEevZV1nmaO2RxJssWjg"; \
		echo "  💡 Set with: export PORTAL_API_KEY=your_api_key"; \
		exit 1; \
	fi
	@echo "✅ Environment variables are set"
	@echo ""
	@echo "⚙️  Configuring for LOCAL PATH..."
	@if [[ "$$(uname)" == "Darwin" ]]; then \
		sed -i '' 's/portal_application_id: ""/portal_application_id: "$(PORTAL_APPLICATION_ID)"/; s/portal_api_key: ""/portal_api_key: "test_api_key"/' e2e/config/.e2e_load_test.config.yaml; \
	else \
		sed -i 's/portal_application_id: ""/portal_application_id: "$(PORTAL_APPLICATION_ID)"/; s/portal_api_key: ""/portal_api_key: "test_api_key"/' e2e/config/.e2e_load_test.config.yaml; \
	fi
	@echo "✅ Config updated for local PATH"
	@echo ""
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "🚀 Running load test with default service IDs: bsc,eth,pocket,poly,gnosis,xrpl_evm_test"; \
		(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon TEST_SERVICE_IDS=bsc,eth,pocket,poly,gnosis,xrpl_evm_test go test -v -tags=e2e -count=1 -run Test_PATH_E2E); \
	else \
		echo "🚀 Running load test with service IDs: $(filter-out $@,$(MAKECMDGOALS))"; \
		(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E); \
	fi

.PHONY: make_load_test_portal_prod_path
make_load_test_portal_prod_path: ## Run a Shannon load test against production PATH with Portal credentials
	@echo "📋 Copying archival config..."
	@cp e2e/config/.grove.archival.e2e_load_test.config.yaml e2e/config/.e2e_load_test.config.yaml
	@echo "✅ Config copied successfully"
	@echo ""
	@echo "🔍 Checking environment variables..."
	@if [ -z "$(PORTAL_APPLICATION_ID)" ]; then \
		echo "❌ Error: PORTAL_APPLICATION_ID environment variable is not set"; \
		echo "  🔐 Grove employees: Get credentials from https://share.1password.com/s#x4x-cWnkf4GUw50BwV166aWEEevZV1nmaO2RxJssWjg"; \
		echo "  💡 Set with: export PORTAL_APPLICATION_ID=your_app_id"; \
		exit 1; \
	fi
	@if [ -z "$(PORTAL_API_KEY)" ]; then \
		echo "❌ Error: PORTAL_API_KEY environment variable is not set"; \
		echo "  🔐 Grove employees: Get credentials from https://share.1password.com/s#x4x-cWnkf4GUw50BwV166aWEEevZV1nmaO2RxJssWjg"; \
		echo "  💡 Set with: export PORTAL_API_KEY=your_api_key"; \
		exit 1; \
	fi
	@echo "✅ Environment variables are set"
	@echo ""
	@echo "⚙️  Configuring for PRODUCTION PATH..."
	@if [[ "$$(uname)" == "Darwin" ]]; then \
		sed -i '' 's/portal_application_id: ""/portal_application_id: "$(PORTAL_APPLICATION_ID)"/; s/portal_api_key: ""/portal_api_key: "$(PORTAL_API_KEY)"/' e2e/config/.e2e_load_test.config.yaml; \
	else \
		sed -i 's/portal_application_id: ""/portal_application_id: "$(PORTAL_APPLICATION_ID)"/; s/portal_api_key: ""/portal_api_key: "$(PORTAL_API_KEY)"/' e2e/config/.e2e_load_test.config.yaml; \
	fi
	@echo "✅ Config updated for production PATH"
	@echo ""
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "🚀 Running load test with default service IDs: bsc,eth,pocket,poly,gnosis,xrpl_evm_test"; \
		(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon TEST_SERVICE_IDS=bsc,eth,pocket,poly,gnosis,xrpl_evm_test go test -v -tags=e2e -count=1 -run Test_PATH_E2E); \
	else \
		echo "🚀 Running load test with service IDs: $(filter-out $@,$(MAKECMDGOALS))"; \
		(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E); \
	fi

# Allow service IDs to be passed as arguments
%:
	@:

##################
### Morse Tests ###
##################

# Targets are also provided to run a morse load test, which use the `morse_load_test` targets

.PHONY: morse_load_test_all
morse_load_test_all: morse_e2e_config_warning ## Run a Morse load test for all service IDs
	(cd e2e && TEST_MODE=load TEST_PROTOCOL=morse go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

.PHONY: morse_load_test
morse_load_test: morse_e2e_config_warning ## Run a Morse load test with specified service IDs (e.g. make morse_load_test F00C,F021)
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "❌ Error: Service IDs are required (comma-separated list)"; \
		echo "  👀 Example: make morse_load_test F00C,F021"; \
		echo "  💡 To run with default service IDs, use: make morse_load_test_defaults"; \
		exit 1; \
	fi
	@(cd e2e && TEST_MODE=load TEST_PROTOCOL=morse TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

.PHONY: copy_e2e_load_test_config
copy_e2e_load_test_config: ## Copy the e2e_load_test.config.tmpl.yaml to e2e_load_test.config.yaml and configure Portal credentials
	@./e2e/scripts/copy_load_test_config.sh

# In order to allow passing the service IDs to the load test targets, this target is needed to avoid printing an error.
%:
	@:
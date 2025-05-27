# DEV_NOTE: DO NOT CHANGE the (cd e2e && go test ...) to the (go test ... e2e)
# in the helpers below. This is needed to ensure the logs are beautified as expected.

#########################
### Test Make Targets ###
#########################

.PHONY: test_all ## Run all tests
test_all: test_unit test_e2e_evm_shannon_defaults test_e2e_evm_morse_defaults

.PHONY: test_unit
test_unit: ## Run all unit tests
	go test ./... -short -count=1

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
	(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

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
	(cd e2e && TEST_MODE=load TEST_PROTOCOL=morse TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E)

.PHONY: copy_e2e_load_test_config
copy_e2e_load_test_config: ## Copy the e2e_load_test.config.tmpl.yaml to e2e_load_test.config.yaml and configure Portal credentials
	@./e2e/scripts/copy_load_test_config.sh

# In order to allow passing the service IDs to the load test targets, this target is needed to avoid printing an error.
%:
	@:
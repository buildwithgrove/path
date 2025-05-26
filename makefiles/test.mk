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

.PHONY: test_e2e_evm_morse_defaults
test_e2e_evm_morse_defaults: morse_e2e_config_warning ## Run an E2E Morse relay test with default service IDs
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=morse go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: test_e2e_evm_morse
test_e2e_evm_morse: morse_e2e_config_warning ## Run an E2E Morse relay test with specified service IDs (e.g. make test_e2e_evm_morse F00C,F021)
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "❌ Error: Service IDs are required (comma-separated list)"; \
		echo "  👀 Example: make test_e2e_evm_morse F00C,F021"; \
		echo "  💡 To run with default service IDs, use: make test_e2e_evm_morse_defaults"; \
		exit 1; \
	fi
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=morse TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: test_e2e_evm_shannon_defaults
test_e2e_evm_shannon_defaults: shannon_e2e_config_warning ## Run an E2E Shannon relay test with default service IDs
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=shannon go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: test_e2e_evm_shannon
test_e2e_evm_shannon: shannon_e2e_config_warning ## Run an E2E Shannon relay test with specified service IDs (e.g. make test_e2e_evm_shannon eth,anvil)
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "❌ Error: Service IDs are required (comma-separated list)"; \
		echo "  👀 Example: make test_e2e_evm_shannon eth,anvil"; \
		echo "  💡 To run with default service IDs, use: make test_e2e_evm_shannon_defaults"; \
		exit 1; \
	fi
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=shannon TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

##################
### Load Tests ###
##################

.PHONY: test_load_evm_morse_defaults
test_load_evm_morse_defaults: morse_e2e_config_warning ## Run a Morse load test with default service IDs
	(cd e2e && TEST_MODE=load TEST_PROTOCOL=morse go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: test_load_evm_morse
test_load_evm_morse: morse_e2e_config_warning ## Run a Morse load test with specified service IDs (e.g. make test_load_evm_morse F00C,F021)
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "❌ Error: Service IDs are required (comma-separated list)"; \
		echo "  👀 Example: make test_load_evm_morse F00C,F021"; \
		echo "  💡 To run with default service IDs, use: make test_load_evm_morse_defaults"; \
		exit 1; \
	fi
	(cd e2e && TEST_MODE=load TEST_PROTOCOL=morse TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: test_load_evm_shannon_defaults
test_load_evm_shannon_defaults: shannon_e2e_config_warning ## Run a Shannon load test with default service IDs
	(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: test_load_evm_shannon
test_load_evm_shannon: shannon_e2e_config_warning ## Run a Shannon load test with specified service IDs (e.g. make test_load_evm_shannon eth,anvil)
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" = "" ]; then \
		echo "❌ Error: Service IDs are required (comma-separated list)"; \
		echo "  👀 Example: make test_load_evm_shannon eth,anvil"; \
		echo "  💡 To run with default service IDs, use: make test_load_evm_shannon_defaults"; \
		exit 1; \
	fi
	(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon TEST_SERVICE_IDS=$(filter-out $@,$(MAKECMDGOALS)) go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: copy_e2e_load_test_config
copy_e2e_load_test_config: ## Copy the e2e_load_test.config.tmpl.yaml to e2e_load_test.config.yaml
	@echo "📁 Copying e2e_load_test.config.tmpl.yaml to e2e_load_test.config.yaml"
	cp ./e2e/config/e2e_load_test.config.tmpl.yaml ./e2e/config/.e2e_load_test.config.yaml
	@echo "✅ Successfully copied e2e_load_test.config.tmpl.yaml to e2e_load_test.config.yaml"
	@echo "💡 To customize the load test config, edit the YAML file at ./e2e/config/.e2e_load_test.config.yaml"

# Rule to handle arbitrary targets (service IDs)
%:
	@:

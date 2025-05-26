# DEV_NOTE: DO NOT CHANGE the (cd e2e && go test ...) to the (go test ... e2e)
# in the helpers below. This is needed to ensure the logs are beautified as expected.

#########################
### Test Make Targets ###
#########################

.PHONY: test_all ## Run all tests
test_all: test_unit test_e2e_evm_shannon test_e2e_evm_morse

.PHONY: test_unit
test_unit: ## Run all unit tests
	go test ./... -short -count=1

#################
### E2E Tests ###
#################

.PHONY: test_e2e_evm_morse
test_e2e_evm_morse: morse_e2e_config_warning ## Run an E2E Morse relay test
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=morse go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: test_e2e_evm_shannon
test_e2e_evm_shannon: shannon_e2e_config_warning ## Run an E2E Shannon relay test
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=shannon go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

##################
### Load Tests ###
##################

.PHONY: test_load_evm_morse
test_load_evm_morse: morse_e2e_config_warning ## Run a Morse load test
	(cd e2e && TEST_MODE=load TEST_PROTOCOL=morse go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: test_load_evm_shannon
test_load_evm_shannon: shannon_e2e_config_warning ## Run a Shannon load test
	(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: copy_e2e_load_test_config
copy_e2e_load_test_config:
	@echo "📁 Copying e2e_load_test.config.tmpl.yaml to e2e_load_test.config.yaml"
	cp ./e2e/config/e2e_load_test.config.tmpl.yaml ./e2e/config/.e2e_load_test.config.yaml
	@echo "✅ Successfully copied e2e_load_test.config.tmpl.yaml to e2e_load_test.config.yaml"
	@echo "💡 To customize the load test config, edit the YAML file at ./e2e/config/.e2e_load_test.config.yaml"

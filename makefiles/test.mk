#########################
### Test Make Targets ###
#########################

.PHONY: test_all ## Run all tests
test_all: test_unit test_e2e_evm_shannon test_e2e_evm_morse

.PHONY: test_unit
test_unit: ## Run all unit tests
	go test ./... -short -count=1

.PHONY: test_e2e_evm_morse
test_e2e_evm_morse: morse_e2e_config_warning ## Run an E2E Morse relay test
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=morse go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: test_load_evm_morse
test_load_evm_morse: morse_e2e_config_warning ## Run a Morse load test
	(cd e2e && TEST_MODE=load TEST_PROTOCOL=morse go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: test_e2e_evm_shannon
test_e2e_evm_shannon: shannon_e2e_config_warning ## Run an E2E Shannon relay test
	(cd e2e && TEST_MODE=e2e TEST_PROTOCOL=shannon go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: test_load_evm_shannon
test_load_evm_shannon: shannon_e2e_config_warning ## Run a Shannon load test
	(cd e2e && TEST_MODE=load TEST_PROTOCOL=shannon go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: copy_e2e_config
copy_e2e_config:
	@echo "📁 Copying e2e config template to e2e config file"
	cp ./e2e/config/e2econfig.tmpl.yaml ./e2e/config/.e2econfig.yaml
	@echo "✅ Successfully copied e2e config template to e2e config file"
	@echo "  💡 To customize the e2e config, edit the YAML file at ./e2e/config/.e2econfig.yaml"

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
	TEST_PROTOCOL=morse go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM ./e2e

.PHONY: test_e2e_evm_shannon
test_e2e_evm_shannon: shannon_e2e_config_warning ## Run an E2E Shannon relay test
	TEST_PROTOCOL=shannon go test -v -tags=e2e -count=1 -run Test_PATH_E2E_EVM ./e2e

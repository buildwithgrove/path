#########################
### Test Make Targets ###
#########################

.PHONY: test_all ## Run all tests
test_all: test_unit test_e2e_evm_shannon test_e2e_evm_morse

.PHONY: test_unit
test_unit: ## Run all unit tests
	go test ./... -short -count=1

.PHONY: test_e2e_evm_morse
test_e2e_evm_morse: morse_e2e_config_warning debug_view_results_links ## Run an E2E Morse relay test
	(cd ./e2e && TEST_PROTOCOL=morse go test -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: test_e2e_evm_morse_with_logs
test_e2e_evm_morse_with_logs: morse_e2e_config_warning debug_view_results_links ## Run an E2E Morse relay test with logs saved to ./e2e/logs
	(cd ./e2e && TEST_PROTOCOL=morse DOCKER_LOG=true go test -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: test_e2e_evm_shannon
test_e2e_evm_shannon: shannon_e2e_config_warning debug_view_results_links ## Run an E2E Shannon relay test
	(cd ./e2e && TEST_PROTOCOL=shannon go test -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

.PHONY: test_e2e_evm_shannon_with_logs
test_e2e_evm_shannon_with_logs: shannon_e2e_config_warning debug_view_results_links ## Run an E2E Shannon relay test with logs saved to ./e2e/logs
	(cd ./e2e && TEST_PROTOCOL=shannon DOCKER_LOG=true go test -tags=e2e -count=1 -run Test_PATH_E2E_EVM)

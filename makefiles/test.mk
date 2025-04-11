#########################
### Test Make Targets ###
#########################

.PHONY: test_all ## Run all tests
test_all: test_unit test_e2e_shannon_relay test_e2e_morse_relay

.PHONY: test_unit
test_unit: ## Run all unit tests
	go test ./... -short -count=1

.PHONY: test_e2e_morse_relay
test_e2e_morse_relay: morse_e2e_config_warning ## Run an E2E Morse relay test
	go test -v ./e2e/... -tags=e2e -count=1 -run Test_MorseRelay
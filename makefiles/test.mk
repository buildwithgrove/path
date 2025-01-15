#########################
### Test Make Targets ###
#########################

.PHONY: test_all ## Run all tests
test_all: test_unit test_auth_server test_e2e_shannon_relay test_e2e_morse_relay

.PHONY: test_unit
test_unit: ## Run all unit tests
	go test ./... -short -count=1

.PHONY: test_auth_server
test_auth_server: ## Run the auth server tests
	(cd envoy/auth_server && go test ./... -count=1)

.PHONY: test_e2e_shannon_relay_iterate
test_e2e_shannon_relay_iterate: ## Iterate on E2E shannon relay tests
	@echo "go build -o bin/path ./cmd"
	@echo "# Update ./bin/config/.config.yaml"
	@echo "./bin/path"
	@echo "curl http://anvil.localhost:3069/v1/abcd1234 -X POST -H \"Content-Type: application/json\" -d '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"eth_blockNumber\"}'"

.PHONY: test_e2e_shannon_relay
test_e2e_shannon_relay: shannon_e2e_config_warning ## Run an E2E Shannon relay test
	@echo "###############################################################################################################################################################"
	@echo "### README: If you are intending to iterate on E2E tests, stop this and run the following for instructions instead: 'make test_e2e_shannon_relay_iterate'. ###"
	@echo "###############################################################################################################################################################"
	go test -v ./e2e/... -tags=e2e -count=1 -run Test_ShannonRelay

.PHONY: test_e2e_morse_relay
test_e2e_morse_relay: morse_e2e_config_warning ## Run an E2E Morse relay test
	go test -v ./e2e/... -tags=e2e -count=1 -run Test_MorseRelay
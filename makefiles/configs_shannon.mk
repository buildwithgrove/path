
.PHONY: shannon_e2e_config_warning
# Internal helper: Checks for required Shannon E2E test config files
shannon_e2e_config_warning:
	$(call check_config_exists,./e2e/.shannon.config.yaml,shannon_prepare_e2e_config)

.PHONY: shannon_populate_config
shannon_populate_config: ## Populates the shannon config file with the correct values
	./local/scripts/shannon_populate_config.sh

.PHONY: shannon_prepare_e2e_config
shannon_prepare_e2e_config: ## Setup Shannon E2E test config file from the example template
	@if [ ! -f ./e2e/.shannon.config.yaml ]; then \
		cp ./config/examples/config.shannon_example.yaml ./e2e/.shannon.config.yaml; \
		echo "################################################################"; \
		echo "Created ./e2e/.shannon.config.yaml"; \
		echo ""; \
		echo "Next steps:"; \
		echo ""; \
		echo "For external contributors:"; \
		echo "  Update the following values in .shannon.config.yaml:"; \
		echo "    - gateway_private_key_hex"; \
		echo "    - owned_apps_private_keys_hex"; \
		echo ""; \
		echo "For Grove employees:"; \
		echo "  1. Search for 'PATH' in 1Password"; \
		echo "  2. Copy and paste the appropriate config into ./e2e/.shannon.config.yaml"; \
		echo ""; \
		echo "Then, for E2E tests:"; \
		echo "  make test_e2e_shannon_relay"; \
		echo ""; \
		echo "Alternatively, for local development"; \
		echo "  make copy_shannon_e2e_config_to_local"; \
		echo "  make path_up"; \
		echo "################################################################"; \
		echo "################################################################"; \
	else \
		echo "################################################################"; \
		echo "Warning: ./e2e/.shannon.config.yaml already exists"; \
		echo "To recreate the file, delete it first and run this command again"; \
		echo "	rm ./e2e/.shannon.config.yaml"; \
		echo "	make shannon_prepare_e2e_config"; \
		echo "################################################################"; \
	fi
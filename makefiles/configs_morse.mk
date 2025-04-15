.PHONY: morse_e2e_config_warning
## Internal helper: Checks for required Morse E2E test config files
morse_e2e_config_warning:
	$(call check_config_exists,./e2e/.morse.config.yaml,morse_prepare_e2e_config)

.PHONY: morse_prepare_e2e_config
morse_prepare_e2e_config: ## Setup Morse E2E test config file from the example template
	@if [ ! -f ./e2e/.morse.config.yaml ]; then \
		cp ./config/examples/config.morse_example.yaml ./e2e/.morse.config.yaml; \
		echo "################################################################"; \
		echo "Created ./e2e/.morse.config.yaml"; \
		echo ""; \
		echo "Next steps:"; \
		echo ""; \
		echo "For external contributors:"; \
		echo "  Update the following values in .morse.config.yaml:"; \
		echo "  - url"; \
		echo "  - relay_signing_key"; \
		echo "  - signed_aats"; \
		echo ""; \
		echo "For Grove employees:"; \
		echo "  1. Search for 'PATH' in 1Password"; \
		echo "  2. Replace the contents of ./e2e/.morse.config.yaml with the config"; \
		echo ""; \
		echo "Then, for E2E tests:"; \
		echo "  make test_e2e_morse_relay"; \
		echo ""; \
		echo "Alternatively, for local development"; \
		echo "  make morse_copy_e2e_config_to_local"; \
		echo "  make path_up"; \
		echo "################################################################"; \
	else \
		echo "################################################################"; \
		echo "Warning: ./e2e/.morse.config.yaml already exists"; \
		echo "To recreate the file, delete it first and run this command again"; \
		echo "	rm ./e2e/.morse.config.yaml"; \
		echo "	make morse_prepare_e2e_config"; \
		echo "################################################################"; \
	fi

.PHONY: morse_copy_e2e_config_to_local
morse_copy_e2e_config_to_local: ## Copy Morse E2E config to local/path/ directory
	$(call check_config_exists,./e2e/.morse.config.yaml,morse_prepare_e2e_config)
	$(call warn_file_exists,./local/path/.config.yaml)
	@cp ./e2e/.morse.config.yaml ./local/path/.config.yaml
	@echo "################################################################"
	@echo "Successfully copied configuration:"
	@echo "  From: ./e2e/.morse.config.yaml"
	@echo "  To:   ./local/path/.config.yaml"
	@echo "################################################################"
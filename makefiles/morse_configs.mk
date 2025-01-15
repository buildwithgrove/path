.PHONY: morse_e2e_config_warning
morse_e2e_config_warning: ## Checks for required Morse E2E config file
	$(call check_config_exists,./e2e/.morse.config.yaml,copy_morse_e2e_config)

.PHONY: copy_morse_e2e_config
copy_morse_e2e_config: ## Setup Morse E2E test configuration file from example template
	@if [ ! -f ./e2e/.morse.config.yaml ]; then \
		cp ./config/examples/config.morse_example.yaml ./e2e/.morse.config.yaml; \
		echo "################################################################"; \
		echo "Created ./e2e/.morse.config.yaml"; \
		echo ""; \
		echo "Next steps:"; \
		echo ""; \
		echo "For Grove employees:"; \
		echo "For Grove employees:"; \
		echo "  1. Search for 'PATH' in 1Password"; \
		echo "  2. Replace the contents of ./e2e/.shannon.config.yaml with the config"; \
		echo ""; \
		echo "For external contributors:"; \
		echo "  Update the following values in .morse.config.yaml:"; \
		echo "  - url"; \
		echo "  - relay_signing_key"; \
		echo "  - signed_aats"; \
		echo "################################################################"; \
	else \
		echo "################################################################"; \
		echo "Warning: ./e2e/.morse.config.yaml already exists"; \
		echo "To recreate the file, delete it first and run this command again"; \
		echo "################################################################"; \
	fi

.PHONY: copy_morse_e2e_config_to_bin
copy_morse_e2e_config_to_bin: ## Copy Morse E2E config to bin/config directory for binary usage
	$(call check_config_exists,./e2e/.morse.config.yaml,copy_morse_e2e_config)
	@mkdir -p ./bin/config
	$(call warn_file_exists,./bin/config/.config.yaml)
	@cp ./e2e/.morse.config.yaml ./bin/config/.config.yaml
	@echo "################################################################"
	@echo "Successfully copied configuration:"
	@echo "  From: ./e2e/.morse.config.yaml"
	@echo "  To:   ./bin/config/.config.yaml"
	@echo "################################################################"
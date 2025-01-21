.PHONY: install_poktrolld
install_poktrolld: ## Installs the poktrolld binary
	./local/scripts/install_poktrolld_cli.sh

.PHONY: shannon_populate_config
shannon_populate_config: prepare_shannon_e2e_config ## Populates the shannon config file with the correct values
	./local/scripts/shannon_populate_config.sh

.PHONY: shannon_e2e_config_warning
shannon_e2e_config_warning: ## Checks for required Shannon E2E config file
	$(call check_config_exists,./e2e/.shannon.config.yaml,prepare_shannon_e2e_config)

.PHONY: prepare_shannon_e2e_config
prepare_shannon_e2e_config: ## Setup Shannon E2E test configuration file from example template
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
		echo "################################################################"; \
	else \
		echo "################################################################"; \
		echo "Warning: ./e2e/.shannon.config.yaml already exists"; \
		echo "To recreate the file, delete it first and run this command again"; \
		echo "	rm ./e2e/.shannon.config.yaml"; \
		echo "	make prepare_shannon_e2e_config"; \
		echo "################################################################"; \
	fi

.PHONY: copy_shannon_config_to_local
copy_shannon_config_to_local: ## Copy Shannon E2E config to local/path/config directory
	$(call check_config_exists,./e2e/.shannon.config.yaml,prepare_shannon_e2e_config)
	@mkdir -p ./local/path/config
	$(call warn_file_exists,./local/path/config/.config.yaml)
	@cp ./e2e/.shannon.config.yaml ./local/path/config/.config.yaml
	@echo "################################################################"
	@echo "Successfully copied configuration:"
	@echo "  From: ./e2e/.shannon.config.yaml"
	@echo "  To:   ./local/path/config/.config.yaml"
	@echo "################################################################"

.PHONY: copy_shannon_config_to_bin
copy_shannon_config_to_bin: ## Copy Shannon E2E config to bin/config directory for binary usage
	$(call check_config_exists,./e2e/.shannon.config.yaml,prepare_shannon_e2e_config)
	@mkdir -p ./bin/config
	$(call warn_file_exists,./bin/config/.config.yaml)
	@cp ./e2e/.shannon.config.yaml ./bin/config/.config.yaml
	@echo "################################################################"
	@echo "Successfully copied configuration:"
	@echo "  From: ./e2e/.shannon.config.yaml"
	@echo "  To:   ./bin/config/.config.yaml"
	@echo "################################################################"
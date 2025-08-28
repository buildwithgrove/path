.PHONY: shannon_e2e_config_warning
# Internal helper: Checks for required Shannon E2E test config files
shannon_e2e_config_warning:
	$(call check_config_exists,./e2e/config/.shannon.config.yaml,config_prepare_shannon_e2e)

.PHONY: configs_copy_shannon_yaml
configs_copy_shannon_yaml: ## Copies the shannon config file to the local directory.
	@echo "TODO(@olshansk): Create simple configs so new users can get started quickly"
	@exit 1

.PHONY: config_shannon_populate
config_shannon_populate: ## Populates the shannon config file with the correct values
	./local/scripts/shannon_populate_config.sh

.PHONY: config_copy_e2e_load_test
config_copy_e2e_load_test: ## Copy the e2e_load_test.config.default.yaml to e2e_load_test.config.yaml and configure Portal credentials
	@./e2e/scripts/copy_e2e_load_test_config.sh

.PHONY: config_prepare_shannon_e2e
config_prepare_shannon_e2e: ## Setup Shannon E2E test config file from the example template
	@if [ ! -f ./e2e/config/.shannon.config.yaml ]; then \
		cp ./config/examples/config.shannon_example.yaml ./e2e/config/.shannon.config.yaml; \
		echo "################################################################"; \
		echo "Created ./e2e/config/.shannon.config.yaml"; \
		echo ""; \
		echo "Next steps:"; \
		echo ""; \
		echo "👥 For external contributors:"; \
		echo "  - Update in .shannon.config.yaml:"; \
		echo "    • gateway_private_key_hex"; \
		echo "    • owned_apps_private_keys_hex"; \
		echo ""; \
		echo "🌿 For Grove employees:"; \
		echo "  - Search for 'PATH' in 1Password"; \
		echo "  - Copy/paste config into ./e2e/config/.shannon.config.yaml"; \
		echo ""; \
		echo "Then, for E2E tests:"; \
		echo "  make test_e2e_evm_shannon"; \
		echo ""; \
		echo "🧑‍💻 For local dev:"; \
		echo "  make config_shannon_populate"; \
		echo "  make path_up"; \
		echo "################################################################"; \
	else \
		echo "################################################################"; \
		echo "Warning: ./e2e/config/.shannon.config.yaml already exists"; \
		echo "To recreate:"; \
		echo "  rm ./e2e/config/.shannon.config.yaml"; \
		echo "  make config_prepare_shannon_e2e"; \
		echo "################################################################"; \
	fi

.PHONY: config_copy_path_local_config_shannon_e2e
config_copy_path_local_config_shannon_e2e: ## Copy local PATH config to Shannon E2E config
	$(call check_config_exists,./local/path/.config.yaml,config_shannon_populate)
	$(call warn_file_exists,./e2e/config/.shannon.config.yaml)
	@cp ./local/path/.config.yaml ./e2e/config/.shannon.config.yaml
	@echo "################################################################"
	@echo "Successfully copied configuration:"
	@echo "  From: ./local/path/.config.yaml"
	@echo "  To:   ./e2e/config/.shannon.config.yaml"
	@echo "################################################################"
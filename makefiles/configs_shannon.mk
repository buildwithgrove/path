.PHONY: shannon_e2e_config_warning
# Internal helper: Checks for required Shannon E2E test config files
shannon_e2e_config_warning:
	$(call check_config_exists,./e2e/config/.shannon.config.yaml,shannon_prepare_e2e_config)

.PHONY: shannon_prepare_e2e_config
shannon_prepare_e2e_config: ## Setup Shannon E2E test config file from the example template
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
		echo "  make shannon_populate_config OR make shannon_copy_e2e_load_test_config_to_local"; \
		echo "  make path_up"; \
		echo "################################################################"; \
	else \
		echo "################################################################"; \
		echo "Warning: ./e2e/config/.shannon.config.yaml already exists"; \
		echo "To recreate:"; \
		echo "  rm ./e2e/config/.shannon.config.yaml"; \
		echo "  make shannon_prepare_e2e_config"; \
		echo "################################################################"; \
	fi

.PHONY: shannon_copy_e2e_load_test_config_to_local
shannon_copy_e2e_load_test_config_to_local: ## Copy Shannon E2E config to local/path/ directory
	$(call check_config_exists,./e2e/config/.shannon.config.yaml,shannon_prepare_e2e_config)
	$(call warn_file_exists,./local/path/.config.yaml)
	@cp ./e2e/config/.shannon.config.yaml ./local/path/.config.yaml
	@echo "################################################################"
	@echo "Successfully copied configuration:"
	@echo "  From: ./e2e/config/.shannon.config.yaml"
	@echo "  To:   ./local/path/.config.yaml"
	@echo "################################################################"


.PHONY: shannon_populate_config
shannon_populate_config: ## Populates the shannon config file with the correct values
	./local/scripts/shannon_populate_config.sh
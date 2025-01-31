################################################################
# Shared config helpers
################################################################

# Helper function to check if a config file exists
define check_config_exists
	@if [ ! -f $(1) ]; then \
		echo "################################################################"; \
		echo "ERROR: Missing required configuration file"; \
		echo "Required file not found: $(1)"; \
		echo ""; \
		echo "To fix this:"; \
		echo "  Run: make $(2)"; \
		echo "################################################################"; \
		exit 1; \
	fi
endef

# Helper function to warn about existing files
define warn_file_exists
	@if [ -f $(1) ]; then \
		echo "################################################################"; \
		echo "Warning: $(1) already exists"; \
		echo "To override, delete the existing file and run this command again"; \
		echo "################################################################"; \
		exit 1; \
	fi
endef


.PHONY: clear_all_local_configs
clear_all_local_configs: ## Clear all local configs
	rm -f ./bin/config/.config.yaml
	rm -f ./config/.config.yaml
	rm -f ./e2e/.shannon.config.yaml
	rm -f ./e2e/.morse.config.yaml
	rm -f ./local/path/config/.config.yaml
	@echo "################################################################"
	@echo "Cleared all local configs"
	@echo "################################################################"
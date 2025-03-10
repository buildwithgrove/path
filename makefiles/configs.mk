################################################################
# Shared config helpers
################

# Helper variable for all config files
CONFIG_FILES := \
	./bin/config/.config.yaml \
	./config/.config.yaml \
	./local/path/.config.yaml \
	./e2e/.shannon.config.yaml \
	./e2e/.morse.config.yaml \
	./local/path/envoy/.envoy.yaml \
	./local/path/envoy/.ratelimit.yaml \
	./local/path/envoy/.allowed-services.lua \
	./local/path/envoy/.gateway-endpoints.yaml

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

# Helper function to backup and remove a file
define backup_and_remove
	if [ -f $(1) ]; then \
		echo "Processing $(1)..."; \
		cp -f $(1) $(1).backup 2>/dev/null && \
		rm -f $(1) && \
		echo "✓ Backed up and removed $(1)"; \
	else \
		echo "⚠ Skipping $(1) - file does not exist"; \
	fi
endef

.PHONY: check_clear_confirmation
check_clear_confirmation: ## Confirm if user wants to proceed with config cleanup
	@echo "################################################################"
	@echo "WARNING: This will delete the following files:"
	@for file in $(CONFIG_FILES); do \
		echo "- $$file"; \
	done
	@echo "Backups will be created with .backup extension"
	@echo "################################################################"
	@read -p "Are you sure you want to proceed? [y/N] " -n 1 -r; \
	echo; \
	if [[ ! $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "Operation cancelled"; \
		exit 1; \
	fi
	@echo "################################################################"

.PHONY: clear_all_local_configs
clear_all_local_configs: check_clear_confirmation ## Clear all local configs with backup
	@echo "Starting backup and cleanup process..."
	@for file in $(CONFIG_FILES); do \
		eval "$(call backup_and_remove,$$file)"; \
	done
	@echo "################################################################"
	@echo "Completed: All configs processed (backups created where applicable)"
	@echo "################################################################"
################################################################
# Shared config helpers
################

# Helper variable for all config files
CONFIG_FILES := \
	./bin/config/.config.yaml \
	./config/.config.yaml \
	./local/path/.config.yaml \
	./e2e/config/.shannon.config.yaml \
	./e2e/config/.morse.config.yaml \
	./e2e/config/.e2e_load_test.config.yaml \
	./local/path/envoy/.envoy.yaml

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
## Internal helper: Confirm if user wants to proceed with config cleanup
check_clear_confirmation:
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
		echo "Operation canceled"; \
		exit 1; \
	fi
	@echo "################################################################"

.PHONY: configs_clear_all_local
configs_clear_all_local: check_clear_confirmation ## Clear all local configs with backup
	@echo "Starting backup and cleanup process..."
	@for file in $(CONFIG_FILES); do \
		eval "$(call backup_and_remove,$$file)"; \
	done
	@echo "################################################################"
	@echo "Completed: All configs processed (backups created where applicable)"
	@echo "################################################################"

.PHONY: configs_copy_values_yaml
configs_copy_values_yaml: ## Copies the values template file to the local directory.
	@if [ ! -f ./local/path/.values.yaml ]; then \
		cp ./local/path/values.tmpl.yaml ./local/path/.values.yaml; \
		echo "################################################################"; \
		echo "Created ./local/path/.values.yaml"; \
		echo "################################################################"; \
	else \
		echo "################################################################"; \
		echo "Warning: ./local/path/.values.yaml already exists"; \
		echo "To recreate the file, delete it first and run this command again"; \
		echo "	rm ./local/path/.values.yaml"; \
		echo "	make configs_copy_values_yaml"; \
		echo "################################################################"; \
	fi

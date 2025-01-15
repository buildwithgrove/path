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
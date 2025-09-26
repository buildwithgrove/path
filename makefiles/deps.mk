###############################
### Quickstart Make Targets ###
###############################

.PHONY: install_tools
install_tools: ## Checks for missing local development tools and installs them to start a PATH instance in Tilt (Docker, Kind, kubectl, Helm, Tilt, pocketd)
	./local/scripts/install_tools.sh

.PHONY: install_tools_optional
install_tools_optional: ## Checks for and installs optional local development tools (Relay Util, Graphviz, Mockgen)
	./local/scripts/install_tools_optional.sh

.PHONY: check_relay_util
# Internal helper: Checks if relay-util is installed locally
check_relay_util:
	@if ! command -v relay-util &> /dev/null; then \
		echo "####################################################################################################"; \
		echo "Relay Util is not installed."; \
		echo "To use any Relay Util make targets to send load testing requests please install Relay Util with:"; \
		echo "go install github.com/commoddity/relay-util/v2@latest"; \
		echo "####################################################################################################"; \
	fi

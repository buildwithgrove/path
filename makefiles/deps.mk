###############################
### Quickstart Make Targets ###
###############################

.PHONY: install_tools
install_tools: ## Checks for missing local development tools and installs them to start a PATH instance in Tilt (Docker, Kind, kubectl, Helm, Tilt)
	./local/scripts/install_tools.sh

.PHONY: install_optional_tools
install_optional_tools: ## Checks for and installs optional local development tools (Relay Util, Graphviz, Mockgen)
	./local/scripts/install_optional_tools.sh

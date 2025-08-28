###############################
### Quickstart Make Targets ###
###############################

.PHONY: install_tools
install_tools: ## Checks for missing local development tools and installs them to start a PATH instance in Tilt (Docker, Kind, kubectl, Helm, Tilt, pocketd)
	./local/scripts/install_tools.sh

.PHONY: install_tools_optional
install_tools_optional: ## Checks for and installs optional local development tools (Relay Util, Graphviz, Mockgen)
	./local/scripts/install_tools_optional.sh
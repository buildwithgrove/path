###############################
### Quickstart Make Targets ###
###############################

.PHONY: install_deps
install_deps: install_tools install_pocketd ## Installs all dependencies to start a PATH instance in Tilt

.PHONY: install_tools
install_tools: ## Installs the supporting tools to start a PATH instance in Tilt
	./local/scripts/install_tools.sh

.PHONY: install_pocketd
install_pocketd: ## Installs the pocketd binary
	./local/scripts/install_pocketd_cli.sh
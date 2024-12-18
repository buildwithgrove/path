###############################
### Quickstart Make Targets ###
###############################

.PHONY: install_deps
install_deps: install_tools install_poktrolld ## Installs all dependencies to start a PATH instance in Tilt

.PHONY: install_tools
install_tools: ## Installs the supporting tools to start a PATH instance in Tilt
	./local/scripts/install_tools.sh

.PHONY: install_poktrolld
install_poktrolld: ## Installs the poktrolld binary
	./local/scripts/install_poktrolld_cli.sh

.PHONY: shannon_populate_config
shannon_populate_config: ## Populates the shannon config file with the correct values
	./local/scripts/shannon_populate_config.sh

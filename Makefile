########################
### Makefile Helpers ###
########################

.PHONY: list
list: ## List all make targets
	@${MAKE} -pRrn : -f $(MAKEFILE_LIST) 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | sort

.PHONY: help
.DEFAULT_GOAL := help
help: ## Prints all the targets in all the Makefiles
	@grep -h -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-60s\033[0m %s\n", $$1, $$2}'

#############################
#### PATH Build Targets   ###
#############################

# tl;dr Quick testing & debugging of PATH as a standalone
# This section is intended to just build and run the PATH binary.
# It mimics an E2E real environment.

.PHONY: path_build
path_build: ## Build the path binary locally (does not run anything)
	go build -o bin/path ./cmd

.PHONY: check_path_config
check_path_config: ## Verify that path configuration file exists
	@if [ ! -f ./local/path/.config.yaml ]; then \
		echo "################################################################"; \
   		echo "Error: Missing config file at ./local/path/.config.yaml"; \
   		echo ""; \
   		echo "Initialize using either:"; \
   		echo "  make prepare_shannon_e2e_config"; \
   		echo "  make prepare_morse_e2e_config "; \
   		echo "################################################################"; \
   		exit 1; \
   fi

# The PATH config value can be set via the CONFIG_PATH env variable and defaults to ./local/path/.config.yaml
CONFIG_PATH ?= ../local/path/.config.yaml

.PHONY: path_run
path_run: path_build check_path_config ## Run the path binary as a standalone binary
	(cd bin; ./path -config ${CONFIG_PATH})

#################################
###  Local PATH make targets  ###
#################################

# tl;dr Mimic an E2E real environment.
# This section is intended to spin up and develop a full modular stack that includes
# PATH, Envoy Proxy, Rate Limiter, Auth Server, and any other dependencies.

.PHONY: path_up
path_up: check_path_config dev_up ## Brings up local Tilt development environment which includes PATH and all related dependencies (using kind cluster)
	tilt up

.PHONY: path_down
path_down: dev_down ## Tears down local Tilt development environment which includes PATH and all related dependencies (using kind cluster)

.PHONY: path_help
path_help: ## Prints help commands if you cannot start path
	@echo "################################################################";
	@echo "If you're hitting issues running PATH, try running following commands:";
	@echo "	make path_down";
	@echo "	make path_up";
	@echo "################################################################";


###############################
###    Makefile imports     ###
###############################

# TODO_IMPROVE(@commoddity): Add a target similar to "make docs_update_gov_params_page" in poktroll
# that converts "config/service_qos.go" into markdown documentation.

include ./makefiles/configs.mk
include ./makefiles/deps.mk
include ./makefiles/docs.mk
include ./makefiles/guard.mk
include ./makefiles/localnet.mk
include ./makefiles/morse_configs.mk
include ./makefiles/shannon_configs.mk
include ./makefiles/test.mk
include ./makefiles/test_requests.mk
include ./makefiles/proto.mk
include ./makefiles/debug.mk





 3937  make claudesync_push
 4030  claudesync project init --new --name path_docs --description "PATH documentation" --local-path .
 4033  claudesync project init --new --name path_docs --description "PATH documentation" --local-path ./docusaurus/
 4034  claudesync push
 4037  claudesync category ls
 4038  claudesync config category add --description markdown_files --patterns "*.md" markdown
 4042  claudesync config category list
 4043  claudesync config category ls
 4044  claudesync push --category markdown



# Docusaurus project files
docusaurus/*.js
docusaurus/*.json
docusaurus/*.lock
docusaurus/node_modules
docusaurus/scripts
docusaurus/src
docusaurus/static
docusaurus/yarn.lock

# Non-relevant markdown files
*.md
# Keep important markdown files
!docusaurus/docs/**/*.md
!docusaurus/README.md
!README.md
!**/README.md

# Generated files
docusaurus/.docusaurus
docusaurus/.cache-loader
docusaurus/build

# Misc
.DS_Store
.env*
.idea
.vscode

# Logs
npm-debug.log*
yarn-debug.log*
yarn-error.log*

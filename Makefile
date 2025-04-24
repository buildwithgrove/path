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

.PHONY: path_build_image
path_build_image: ## Builds the PATH Docker development image
	@echo "🔨 Building PATH Docker image..."
	@docker build -t ghcr.io/buildwithgrove/path-localnet-env:latest -f local/Dockerfile.dev .

.PHONY: path_push_image
path_push_image: ## Pushes the PATH Docker image to GitHub Container Registry
	@echo "📦 Pushing PATH Docker image to GitHub Container Registry..."
	@docker push ghcr.io/buildwithgrove/path-localnet-env:latest

.PHONY: path_up
path_up: ## Brings up local Tilt development environment in Docker 
	@./local/scripts/localnet.sh up

.PHONY: path_down
path_down: ## Tears down local Tilt development environment in Docker
	@./local/scripts/localnet.sh down

.PHONY: path_help
path_help: ## Prints help commands if you cannot start path
	@echo "################################################################";
	@echo "💡 If you're hitting issues running PATH, try running following commands:";
	@echo "	make path_down";
	@echo "	make path_up";
	@echo "################################################################";

###############################
###    Makefile imports     ###
###############################

include ./makefiles/configs.mk
include ./makefiles/configs_morse.mk
include ./makefiles/configs_shannon.mk
include ./makefiles/deps.mk
include ./makefiles/docs.mk
include ./makefiles/test.mk
include ./makefiles/test_requests.mk
include ./makefiles/proto.mk
include ./makefiles/debug.mk
include ./makefiles/claude.mk

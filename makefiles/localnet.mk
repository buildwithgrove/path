#################################
###  Local PATH make targets  ###
#################################

# These targets are used to bring up the local Tilt environment in a 
# dedicated Docker container that contains all dependencies for local
# development (Tilt, Helm, etc).
#
# The localnet.sh script handles all the complexity of bringing up the PATH
# services in the Docker container, including checking for the presence of
# the config.yaml and .values.yaml files.
#
# For more information see the documentation at:
# https://path.grove.city/develop/path/environment

# Brings up local Tilt environment with remote helm charts
.PHONY: path_up
path_up: check_docker ## Brings up local Tilt development environment in Docker 
	@./local/scripts/localnet.sh up 

# Brings up local Tilt environment with local helm charts
.PHONY: path_up_local_helm
path_up_local_helm: check_docker ## Brings up local Tilt environment with local helm charts
	@./local/scripts/localnet.sh up --use-local-helm

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

.PHONY: build_and_push_localnet_image
build_and_push_localnet_image: ## Builds and pushes the localnet Docker image for multi-architecture builds
	@echo "🔨 Building and pushing multi-architecture localnet Docker image..."
	@docker buildx build --no-cache --platform linux/amd64,linux/arm64 \
	  -t ghcr.io/buildwithgrove/path-localnet-env:latest \
	  -f ./local/Dockerfile.dev \
	  --push .

#################################
###  Local PATH make targets  ###
#################################

# tl;dr Mimic an E2E real environment.
# This section is intended to spin up and develop a full modular stack

.PHONY: path_up
path_up: check_docker check_path_config ## Brings up local Tilt development environment in Docker 
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

.PHONY: build_and_push_localnet_image
build_and_push_localnet_image: ## Builds and pushes the localnet Docker image for multi-architecture builds
	@echo "🔨 Building and pushing multi-architecture localnet Docker image..."
	@docker buildx build --platform linux/amd64,linux/arm64 \
	  -t ghcr.io/buildwithgrove/path-localnet-env:latest \
	  -f ./local/Dockerfile.dev \
	  --push .

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


.PHONY: check_docker
# Internal helper: Check if Docker is installed locally
check_docker:
	@if ! command -v docker >/dev/null 2>&1; then \
		echo "${RED}Docker is not installed. Please install Docker and try again.${RESET}"; \
		exit 1; \
	fi;
	@if ! docker info >/dev/null 2>&1; then \
		echo "${RED}Docker daemon is not running. Please start Docker and try again.${RESET}"; \
		echo "You can start Docker by doing ONE OF the following:"; \
		echo "  - Opening Docker Desktop application"; \
		echo "  - Running ${CYAN}'sudo systemctl start docker' on Linux${RESET}"; \
		echo "  - Running ${CYAN}'open /Applications/Docker.app' on macOS${RESET}"; \
		exit 1; \
	fi;

.PHONY: check_path_up
# Internal helper: Checks if PATH is running at localhost:3070
check_path_up:
	@if ! nc -z localhost 3070 2>/dev/null; then \
		echo "########################################################################"; \
		echo "ERROR: PATH is not running on port 3070"; \
		echo "Please start it with:"; \
		echo "  make path_up"; \
		echo "########################################################################"; \
		exit 1; \
	fi

.PHONY: path_up
path_up: check_docker ## Brings up local Tilt development environment in Docker with remote helm charts
	@./local/scripts/localnet.sh up

.PHONY: path_up_local_helm
path_up_local_helm: check_docker ## Brings up local Tilt development environment in Docker with local helm charts
	@./local/scripts/localnet.sh up --use-local-helm

.PHONY: path_down
path_down: ## Tears down local Tilt development environment in Docker
	@./local/scripts/localnet.sh down

.PHONY: localnet_exec
localnet_exec: ## Opens a terminal inside the path-localnet container
	@docker exec -it path-localnet /bin/bash

.PHONY: localnet_k9s
localnet_k9s: ## Launches k9s inside the path-localnet container for Kubernetes debugging
	@if ! docker ps --format '{{.Names}}' | grep -q "^path-localnet$$"; then \
		echo "❌ Error: path-localnet container is not running."; \
		echo "Start it first with: make path_up"; \
		exit 1; \
	fi
	@echo "🚀 Launching k9s inside path-localnet container..."
	@docker exec -it path-localnet k9s

.PHONY: build_and_push_localnet_image
build_and_push_localnet_image: ## Builds and pushes the localnet Docker image for multi-architecture builds
	@echo "🔨 Building and pushing multi-architecture localnet Docker image..."
	@docker buildx build --no-cache --platform linux/amd64,linux/arm64 \
	  -t ghcr.io/buildwithgrove/path-localnet-env:latest \
	  -f ./local/Dockerfile.dev \
	  --push .
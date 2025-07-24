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
	
.PHONY: check_docker
# Internal helper: Check if Docker is installed locally
check_docker:
	@if ! command -v docker >/dev/null 2>&1; then \
		echo "Docker is not installed. Make sure you review README.md before continuing"; \
		exit 1; \
	fi;
	@if ! docker info >/dev/null 2>&1; then \
		echo "Docker daemon is not running. Please start Docker and try again."; \
		echo "You can start Docker by doing one of the following:"; \
		echo "  - Opening Docker Desktop application"; \
		echo "  - Running 'sudo systemctl start docker' on Linux"; \
		echo "  - Running 'open /Applications/Docker.app' on macOS"; \
		exit 1; \
	fi;

.PHONY: path_down
path_down: ## Tears down local Tilt development environment in Docker
	@./local/scripts/localnet.sh down

.PHONY: localnet_exec
localnet_exec: ## Opens a terminal inside the path-localnet container
	@docker exec -it path-localnet /bin/bash
.PHONY: k8s_prepare_local_env

# Internal helper for path localnet: creates a kind cluster and namespaces if they don't already exist
k8s_prepare_local_env: check_kind
	@if ! kind get clusters | grep -q "^path-localnet$$"; then \
		echo "[INFO] Cluster 'path-localnet' not found. Creating it..."; \
		kind create cluster --name path-localnet --config ./local/kind-config.yaml; \
		kubectl create namespace path; \
		kubectl create namespace monitoring; \
		kubectl create namespace middleware; \
		kubectl config set-context --current --namespace=path; \
		kubectl create secret generic path-config --from-file=./local/path/.config.yaml -n path; \
	else \
		echo "[DEBUG] Cluster 'path-localnet' already exists. Checking context..."; \
		if ! kubectl config get-contexts | grep -q "^[* ]*kind-path-localnet"; then \
			echo "[INFO] Context 'kind-path-localnet' not found. Setting up kubeconfig..."; \
			kind export kubeconfig --name path-localnet; \
		fi; \
		if ! kubectl get namespace path >/dev/null 2>&1; then \
			echo "[INFO] Creating missing namespaces..."; \
			kubectl create namespace path; \
			kubectl create namespace monitoring; \
			kubectl create namespace middleware; \
			kubectl config set-context --current --namespace=path; \
			kubectl create secret generic path-config --from-file=./local/path/.config.yaml -n path; \
		fi; \
	fi; \
	kubectl config use-context kind-path-localnet;

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

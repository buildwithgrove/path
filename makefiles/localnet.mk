###############################
###  Localnet check targets ###
###############################

.PHONY: check_kind
# Internal helper: Checks if Kind is installed locally
check_kind:
	@if ! command -v kind >/dev/null 2>&1; then \
		echo "kind is not installed. Make sure you review README.md before continuing"; \
		exit 1; \
	fi

.PHONY: check_tilt
# Internal helper: Checks if Tilt is installed locally
check_tilt:
	@if ! command -v tilt >/dev/null 2>&1; then \
		echo "Tilt is not installed. Make sure you review README.md before continuing"; \
		exit 1; \
	fi

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

###############################
### Localnet config targets ###
###############################

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
		echo "[DEBUG] Cluster 'path-localnet' already exists. Skipping creation."; \
	fi; \
	kubectl config use-context kind-path-localnet;

.PHONY: k8s_cleanup_local_env
# Internal helper: Cleans up kind cluster and kubeconfig context for path-localnet
k8s_cleanup_local_env:
	@echo "[INFO] Cleaning up local k8s environment for 'path-localnet'..."
	@kind delete cluster --name path-localnet || echo "[DEBUG] Cluster 'path-localnet' not found. Skipping deletion."
	@kubectl config get-contexts kind-path-localnet > /dev/null 2>&1 && \
		kubectl config delete-context kind-path-localnet || \
		echo "[DEBUG] Context 'kind-path-localnet' not found. Skipping deletion."
	@kubectl config get-contexts | grep -q 'kind-path-localnet' || echo "[INFO] Cleanup complete."
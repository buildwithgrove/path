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

###############################
### Localnet config targets ###
###############################

.PHONY: dev_up
# Internal helper: Spins up Kind cluster if it doesn't already exist
dev_up: check_kind
	@if ! kind get clusters | grep -q "^path-localnet$$"; then \
		echo "[INFO] Cluster 'path-localnet' not found. Creating it..."; \
		kind create cluster --name path-localnet; \
		kubectl config use-context kind-path-localnet; \
	else \
		echo "[DEBUG] Cluster 'path-localnet' already exists. Skipping creation."; \
	fi

# TODO_UPNEXT(@HebertCL): Integrate this process into the Tiltfile to enable hot reloading on the .config.yaml file values to work in Tilt.
.PHONY: config_path_secrets
# Internal helper: Creates path config secret if it doesn't already exist
config_path_secrets: check_path_config
	@echo "[INFO] Checking if secret 'path-config-local' exists..."
	@if ! kubectl get secret path-config-local > /dev/null 2>&1; then \
		echo "[INFO] Secret 'path-config-local' not found. Creating it..."; \
		kubectl create secret generic path-config-local \
			--from-file=.config.yaml=./local/path/config/.config.yaml; \
	else \
		echo "[DEBUG] Secret 'path-config-local' already exists. Skipping creation."; \
	fi

.PHONY: dev_down
# Internal helper: Tears down kind cluster
dev_down:
	@echo "Tearing down local environment..."
	@tilt down
	@kubectl delete secret path-config-local
	@kind delete cluster --name path-localnet
	@if kubectl config get-contexts kind-path-localnet > /dev/null 2>&1; then \
		kubectl config delete-context kind-path-localnet; \
	else \
		echo "Context kind-path-localnet not found in kubeconfig. Skipping deletion."; \
	fi

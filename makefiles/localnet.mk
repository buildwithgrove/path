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

.PHONY: check_path_config_file
# Internal helper: Check if .config.yaml exists
check_path_config:
	@if ! test -f local/path/config/.config.yaml; then \
		echo ".config.yaml file does not exists. Make sure to review README.md and run copy_shannon_config/copy_morse_config targets first"; \
		exit 1; \
	fi

###############################
### Localnet config targets ###
###############################

.PHONY: dev_up
# Internal helper: Spins up Kind cluster
dev_up: check_kind
	@echo "Spinning up local K8s..."
	@kind create cluster --name path-localnet
	@kubectl config use-context kind-path-localnet

.PHONY: config_path_secrets
# Internal helper: Creates a K8s secret based on the .config.yaml file created by copy_shannon_config/copy_morse_config
config_path_secrets: check_path_config
	@echo "Creating path config secret..."
	@kubectl create secret generic path-config-local \
		--from-file=.config.yaml=./local/path/config/.config.yaml

.PHONY: dev_down
# Internal helper: Tears down kind cluster
dev_down:
	@echo "Tearing down local environment..."
	@tilt down
	@kuebctl delete secret path-config-local
	@kind delete cluster --name kind-path-localnet
	@kubectl config delete-context kind-path-localnet
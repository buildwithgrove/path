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
### Run Path Make Targets ###
#############################

.PHONY: path_up
path_up: ## Run docker compose up
	docker compose up -d --no-deps path_gateway

.PHONY: path_up_build
path_up_build: ## Run docker compose up with build
	docker compose up -d --build --no-deps path_gateway

.PHONY: path_up_auth
path_up_auth: ## Run docker compose up with auth
	docker compose up -d

.PHONY: path_up_build_auth
path_up_build_auth: ## Run docker compose up with build and auth
	docker compose up -d --build

.PHONY: path_down
path_down: ## Run docker compose down
	docker compose down

#########################
### Test Make Targets ###
#########################

.PHONY: test_all ## Run all tests
test_all: test_unit test_e2e_shannon_relay

.PHONY: test_unit
test_unit: ## Run all unit tests
	go test ./... -short -count=1
  
.PHONY: test_e2e_shannon_relay
test_e2e_shannon_relay: ## Run an E2E shannon relay test
	go test ./... -tags=e2e -count=1 -run TestShannonRelay 

################################
### Copy Config Make Targets ###
################################

.PHONY: copy_config
copy_config: ## copies the example configuration yaml file to .gitignored .config.yaml file
	@if [ ! -f ./cmd/.config.yaml ]; then \
		cp ./cmd/.config.example.yaml ./cmd/.config.yaml; \
	else \
		echo ".config.yaml already exists, not overwriting."; \
	fi

.PHONY: copy_test_config
copy_test_config: ## copies the example test configuration yaml file to .gitignored .config.test.yaml file
	@if [ ! -f ./e2e/.config.test.yaml ]; then \
		cp ./e2e/.example.test.yaml ./e2e/.config.test.yaml; \
	else \
		echo ".config.test.yaml already exists, not overwriting."; \
	fi

#######################
### SQLC Generation ###
#######################

.PHONY: sqlc_generate
sqlc_generate: ## Generate SQLC code from db/driver/sqlc/*.sql files
	sqlc generate -f ./db/driver/sqlc/sqlc.yaml
	
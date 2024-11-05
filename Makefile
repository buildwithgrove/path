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

# TODO_IMPROVE: add a make target to generate mocks for all the interfaces in the project

#############################
### Run Path Make Targets ###
#############################
.PHONY: path_build
path_build: ## build the path binary
	go build -o bin/path ./cmd

.PHONY: path_up_gateway
path_up_gateway: ## Run just the PATH gateway without any dependencies
	docker compose up -d --no-deps path_gateway

.PHONY: path_up_build_gateway
path_up_build_gateway: ## Run and build just the PATH gateway without any dependencies
	docker compose up -d --build --no-deps path_gateway

# TODO_UPNEXT(@adshmh): update path_up and path_down to use Tilt, and remove docker compose
.PHONY: path_up
path_up: ## Run the PATH gateway and all related dependencies
	docker compose up -d

.PHONY: path_up_build
path_up_build: ## Run and build the PATH gateway and all related dependencies
	docker compose up -d --build

.PHONY: path_down
path_down: ## Stop the PATH gateway and all related dependencies
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
	go test ./... -tags=e2e -count=1 -run Test_ShannonRelay

.PHONY: test_e2e_morse_relay
test_e2e_morse_relay: ## Run an E2E Morse relay test
	go test ./... -tags=e2e -count=1 -run Test_MorseRelay

################################
### Copy Config Make Targets ###
################################

.PHONY: copy_shannon_config
copy_shannon_config: ## copies the example shannon configuration yaml file to .config.yaml file
	@if [ ! -f ./cmd/.config.yaml ]; then \
		cp ./cmd/.config.shannon_example.yaml ./cmd/.config.yaml; \
	else \
		echo ".config.yaml already exists, not overwriting."; \
	fi

.PHONY: copy_morse_config
copy_morse_config: ## copies the example morse configuration yaml file to .config.yaml file
	@if [ ! -f ./cmd/.config.yaml ]; then \
		cp ./cmd/.config.morse_example.yaml ./cmd/.config.yaml; \
	else \
		echo ".config.yaml already exists, not overwriting."; \
	fi

.PHONY: copy_shannon_e2e_config
copy_shannon_e2e_config: ## copies the example Shannon test configuration yaml file to .gitignored .shannon.config.yaml file
	@if [ ! -f ./e2e/.shannon.config.yaml ]; then \
		cp ./e2e/shannon.example.yaml ./e2e/.shannon.config.yaml; \
	else \
		echo "./e2e/.shannon.config.yaml already exists, not overwriting."; \
	fi

.PHONY: copy_morse_e2e_config
copy_morse_e2e_config: ## copies the example Morse test configuration yaml file to .gitignored ..morse.config.yaml file.
	@if [ ! -f ./e2e/.morse.config.yaml ]; then \
		cp ./e2e/morse.example.yaml ./e2e/.morse.config.yaml; \
	else \
		echo "./e2e/.morse.config.yaml already exists, not overwriting."; \
	fi

###############################
### Generation Make Targets ###
###############################

.PHONY: sqlc_generate
sqlc_generate: ## Generate SQLC code from db/driver/sqlc/*.sql files
	sqlc generate -f ./db/driver/sqlc/sqlc.yaml

# // TODO_TECHDEBT(@commoddity): move all mocks to a shared mocks package
# // TODO_TECHDEBT(@commoddity): Add all other mock generation commands here

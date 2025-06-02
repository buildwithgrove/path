#############################
### Devtools Make Targets ###
#############################

# Default service ID if not specified
SERVICE_ID ?= base

# Target to check disqualified endpoints for a specific service
# Usage: make disqualified_endpoints SERVICE_ID=base
# or:    make disqualified_endpoints (uses default SERVICE_ID=base)
.PHONY: disqualified_endpoints
disqualified_endpoints: check_jq
	@echo "🔎 Fetching disqualified endpoints for service: $(SERVICE_ID)"
	curl http://localhost:3069/disqualified_endpoints -H "Target-Service-Id: $(SERVICE_ID)" | jq

.PHONY: check_jq
check_jq: ## Checks if jq is installed
	@if ! command -v jq &> /dev/null; then \
		echo "🚨 jq could not be found. Please install it using your package manager."; \
		exit 1; \
	fi
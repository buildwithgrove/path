#############################
### Devtools Make Targets ###
#############################

# Default service ID if not specified
SERVICE_ID ?= eth

# Target to check disqualified endpoints for a specific service
# Usage: make get_disqualified_endpoints SERVICE_ID=base
# or:    make get_disqualified_endpoints (uses default SERVICE_ID=eth)
.PHONY: get_disqualified_endpoints
get_disqualified_endpoints: check_jq ## Fetch disqualified endpoints for a specific service
	@echo "🔎 Fetching disqualified endpoints for service: $(SERVICE_ID)"
	curl http://localhost:3069/disqualified_endpoints -H "Target-Service-Id: $(SERVICE_ID)" | jq

# DEV_NOTE: in prod the Grove Portal's GUARD will assign the service ID from the subdomain to the `Target-Service-Id` header.
# Usage: make grove_get_disqualified_endpoints SERVICE_ID=base
# or:    make grove_get_disqualified_endpoints (uses default SERVICE_ID=eth)
.PHONY: grove_get_disqualified_endpoints
grove_get_disqualified_endpoints: check_jq ## Fetch disqualified endpoints for Grove's Portal
	@echo "🔎 Fetching disqualified endpoints for Grove's $(SERVICE_ID) service"
	curl https://$(SERVICE_ID).rpc.grove.city/disqualified_endpoints | jq

.PHONY: check_jq
check_jq: # Checks if jq is installed
	@if ! command -v jq &> /dev/null; then \
		echo "🚨 jq could not be found. Please install it using your package manager."; \
		exit 1; \
	fi

#############################
### Devtools Make Targets ###
#############################

# Default Debug API Key if not specified.
DEBUG_API_KEY ?= test_debug_api_key

# Target to check disqualified endpoints for a specific service
# Usage: make get_disqualified_endpoints SERVICE_ID=base (SERVICE_ID is required)
.PHONY: get_disqualified_endpoints
get_disqualified_endpoints: check_jq check_service_id ## Fetch disqualified endpoints for a specific service
	@echo "🔎 Fetching disqualified endpoints for service: $(SERVICE_ID)"
	curl http://localhost:3069/disqualified_endpoints \
		-H "Authorization: $(DEBUG_API_KEY)" \
		-H "Target-Service-Id: $(SERVICE_ID)" | jq

# DEV_NOTE: in prod the Grove Portal's GUARD will assign the service ID from the subdomain to the `Target-Service-Id` header.
# Usage: make grove_get_disqualified_endpoints SERVICE_ID=base
# or:    make grove_get_disqualified_endpoints (uses default SERVICE_ID=eth)
.PHONY: grove_get_disqualified_endpoints
grove_get_disqualified_endpoints: check_jq ## Fetch disqualified endpoints for Grove's Portal
	@echo "🔎 Fetching disqualified endpoints for Grove's $(SERVICE_ID) service"
	curl https://$(SERVICE_ID).rpc.grove.city/disqualified_endpoints | jq

# Health check endpoint
.PHONY: health_check
health_check: ## Check PATH service health
	@echo "🏥 Checking PATH service health..."
	curl http://localhost:3070/healthz

# Performance profiling endpoints
.PHONY: profile_cpu
profile_cpu: ## Capture CPU profile (30 seconds)
	@echo "📊 Capturing CPU profile for 30 seconds..."
	curl http://localhost:3070/debug/pprof/profile?seconds=30 \
		-H "Authorization: $(DEBUG_API_KEY)" \
		-o cpu-$(shell date +%Y%m%d_%H%M%S).prof
	@echo "✅ CPU profile saved as cpu-$(shell date +%Y%m%d_%H%M%S).prof"

.PHONY: profile_memory
profile_memory: ## Capture memory heap profile
	@echo "🧠 Capturing memory heap profile..."
	curl http://localhost:3070/debug/pprof/heap \
		-H "Authorization: $(DEBUG_API_KEY)" \
		-o mem-$(shell date +%Y%m%d_%H%M%S).prof
	@echo "✅ Memory profile saved as mem-$(shell date +%Y%m%d_%H%M%S).prof"

.PHONY: profile_goroutines
profile_goroutines: ## Capture goroutine profile
	@echo "🔄 Capturing goroutine profile..."
	curl http://localhost:3070/debug/pprof/goroutine \
		-H "Authorization: $(DEBUG_API_KEY)" \
		-o goroutine-$(shell date +%Y%m%d_%H%M%S).prof
	@echo "✅ Goroutine profile saved as goroutine-$(shell date +%Y%m%d_%H%M%S).prof"

.PHONY: pprof_index
pprof_index: ## View all available pprof endpoints
	@echo "📋 Available pprof endpoints:"
	curl http://localhost:3070/debug/pprof/ -H "Authorization: $(DEBUG_API_KEY)"

# Convenience target for all monitoring commands (requires SERVICE_ID)
.PHONY: monitor
monitor: check_service_id health_check get_disqualified_endpoints pprof_index ## Run all monitoring checks

.PHONY: check_service_id
check_service_id: ## Check if SERVICE_ID is provided
	@if [ -z "$(SERVICE_ID)" ]; then \
		echo "🚨 SERVICE_ID is required. Usage: make <target> SERVICE_ID=eth"; \
		exit 1; \
	fi

.PHONY: check_jq
check_jq: ## Checks if jq is installed
	@if ! command -v jq &> /dev/null; then \
		echo "🚨 jq could not be found. Please install it using your package manager."; \
		exit 1; \
	fi


###########################
###   Release Helpers   ###
###########################

.PHONY: git_fetch_prune_tags
git_fetch_prune_tags: ## Sync local tags to remote
	git fetch --prune --prune-tags origin

.PHONY: shannon_preliminary_services_test
shannon_preliminary_services_test: ## Run shannon preliminary services test to verify service availability
	@echo "💡 Usage: ./e2e/scripts/shannon_preliminary_services_test.sh [--help]"
	./e2e/scripts/shannon_preliminary_services_test.sh

.PHONY: shannon_services_helpers
shannon_services_helpers: ## Instructions for Shannon service helper functions
	@echo "################################################################"
	@echo "Shannon Service Helper Functions"
	@echo "################################################################"
	@echo "Source the helper functions:"
	@echo "  $$ source ./e2e/scripts/shannon_preliminary_services_helpers.sh"
	@echo ""
	@echo "Available functions:"
	@echo "  $$ shannon_query_services_by_owner --help"
	@echo "  $$ shannon_query_service_tlds_by_id --help"
	@echo "################################################################"
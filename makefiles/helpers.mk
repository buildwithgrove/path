
###########################
###   Release Helpers   ###
###########################

.PHONY: git_fetch_prune_tags
git_fetch_prune_tags: ## Sync local tags to remote
	git fetch --prune --prune-tags origin

.PHONY: shannon_preliminary_services_test_help
shannon_preliminary_services_test_help: ## Run shannon preliminary services test to verify service availability
	./e2e/scripts/shannon_preliminary_services_test.sh --help

.PHONY: shannon_preliminary_services_test
shannon_preliminary_services_test: ## Run shannon preliminary services test to verify service availability
	./e2e/scripts/shannon_preliminary_services_test.sh

.PHONY: source_shannon_preliminary_services_helpers
source_shannon_preliminary_services_helpers: ## Source shannon preliminary services helpers
	@echo "Run the following command to source shannon preliminary services helpers:"
	@echo ""
	@echo "\$$ source ./e2e/scripts/shannon_preliminary_services_helpers.sh"
	@echo ""
	@echo "Then, try running one of the following commands:"
	@echo "	\$$ shannon_query_services_by_owner --help"
	@echo "	\$$ shannon_query_service_tlds_by_id --help"
	@echo ""
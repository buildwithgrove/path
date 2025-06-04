
###########################
###   Release Helpers   ###
###########################

# Sync local tags to remote:
# git fetch --prune --prune-tags origin

.PHONY: shannon_preliminary_services_test
shannon_preliminary_services_test: ## Run shannon preliminary services test to verify service availability
	./e2e/scripts/shannon-preliminary-services-test.sh
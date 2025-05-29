
###########################
###   Release Helpers   ###
###########################

# Sync local tags to remote:
# git fetch --prune --prune-tags origin

.PHONY: release_tag_local_testing
release_tag_local_testing: ## Tag a new local testing release (e.g. v1.0.1 -> v1.0.2-test1, v1.0.2-test1 -> v1.0.2-test2)
	@LATEST_TAG=$$(git tag --sort=-v:refname | head -n 1 | xargs); \
	if [ -z "$$LATEST_TAG" ]; then \
	  NEW_TAG=v0.1.0-test1; \
	else \
	  if echo "$$LATEST_TAG" | grep -q -- '-test'; then \
	    BASE_TAG=$$(echo "$$LATEST_TAG" | sed 's/-test[0-9]*//'); \
	    LAST_TEST_NUM=$$(echo "$$LATEST_TAG" | sed -E 's/.*-test([0-9]+)/\1/'); \
	    NEXT_TEST_NUM=$$(($$LAST_TEST_NUM + 1)); \
	    NEW_TAG=$${BASE_TAG}-test$${NEXT_TEST_NUM}; \
	  else \
	    BASE_TAG=$$(echo "$$LATEST_TAG" | awk -F. -v OFS=. '{$$NF = sprintf("%d", $$NF + 1); print}'); \
	    NEW_TAG=$${BASE_TAG}-test1; \
	  fi; \
	fi; \
	git tag $$NEW_TAG; \
	echo "New local testing version tagged: $$NEW_TAG"; \
	echo "Run the following commands to push the new tag:"; \
	echo "  git push origin $$NEW_TAG"; \
	echo "And draft a new release at https://github.com/buildwithgrove/path/releases/new";


.PHONY: release_tag_dev
release_tag_dev: ## Tag a new dev release (e.g. v1.0.1 -> v1.0.1-dev1, v1.0.1-dev1 -> v1.0.1-dev2)
	@$(eval LATEST_TAG=$(shell git tag --sort=-v:refname | head -n 1))
	@$(eval BASE_VERSION=$(shell echo $(LATEST_TAG) | sed 's/-dev[0-9]*$$//' ))
	@$(eval EXISTING_DEV_TAGS=$(shell git tag --sort=-v:refname | grep "^$(BASE_VERSION)-dev[0-9]*$$" | head -n 1))
	@if [ -z "$(EXISTING_DEV_TAGS)" ]; then \
		NEW_TAG="$(BASE_VERSION)-dev1"; \
	else \
		DEV_NUM=$$(echo $(EXISTING_DEV_TAGS) | sed 's/.*-dev\([0-9]*\)$$/\1/'); \
		NEW_DEV_NUM=$$((DEV_NUM + 1)); \
		NEW_TAG="$(BASE_VERSION)-dev$$NEW_DEV_NUM"; \
	fi; \
	git tag $$NEW_TAG; \
	echo "########"; \
	echo "New dev version tagged: $$NEW_TAG"; \
	echo ""; \
	echo "Next, do the following:"; \
	echo ""; \
	echo "1. Run the following commands to push the new tag:"; \
	echo "   git push origin $$NEW_TAG"; \
	echo ""; \
	echo "2. And draft a new release at https://github.com/buildwithgrove/path/releases/new"; \
	echo ""; \
	echo "If you need to delete a tag, run:"; \
	echo "  git tag -d $$NEW_TAG"; \
	echo ""; \
	echo "If you need to delete a tag remotely, run:"; \
	echo "  git push origin --delete $$NEW_TAG"; \
	echo ""; \
	echo "########"

.PHONY: release_tag_bug_fix
release_tag_bug_fix: ## Tag a new bug fix release (e.g. v1.0.1 -> v1.0.2)
	@$(eval LATEST_TAG=$(shell git tag --sort=-v:refname | head -n 1))
	@LATEST_TAG="$(LATEST_TAG)"; \
	NEW_TAG=$$(echo $$LATEST_TAG | awk -F. -v OFS=. '{ $$NF = sprintf("%d", $$NF + 1); print }'); \
	git tag $$NEW_TAG; \
	echo "########"; \
	echo "New bug fix version tagged: $$NEW_TAG"; \
	echo ""; \
	echo "Next, do the following:"; \
	echo ""; \
	echo "1. Run the following commands to push the new tag:"; \
	echo "   git push origin $$NEW_TAG"; \
	echo ""; \
	echo "2. And draft a new release at https://github.com/buildwithgrove/path/releases/new"; \
	echo ""; \
	echo "If you need to delete a tag, run:"; \
	echo "  git tag -d $$NEW_TAG"; \
	echo ""; \
	echo "If you need to delete a tag remotely, run:"; \
	echo "  git push origin --delete $$NEW_TAG"; \
	echo ""; \
	echo "########"


.PHONY: release_tag_minor_release
release_tag_minor_release: ## Tag a new minor release (e.g. v1.0.0 -> v1.1.0)
	@$(eval LATEST_TAG=$(shell git tag --sort=-v:refname | head -n 1))
	@LATEST_TAG="$(LATEST_TAG)"; \
	NEW_TAG=$$(echo $$LATEST_TAG | awk -F. '{$$2 += 1; $$3 = 0; print $$1 "." $$2 "." $$3}'); \
	git tag $$NEW_TAG; \
	echo "########"; \
	echo "New minor release version tagged: $$NEW_TAG"; \
	echo ""; \
	echo "Next, do the following:"; \
	echo ""; \
	echo "1. Run the following commands to push the new tag:"; \
	echo "   git push origin $$NEW_TAG"; \
	echo ""; \
	echo "2. And draft a new release at https://github.com/buildwithgrove/path/releases/new"; \
	echo ""; \
	echo "If you need to delete a tag, run:"; \
	echo "  git tag -d $$NEW_TAG"; \
	echo ""; \
	echo "If you need to delete a tag remotely, run:"; \
	echo "  git push origin --delete $$NEW_TAG"; \
	echo ""; \
	echo "########";

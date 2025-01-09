#################################
### ClaudeSync make targets ###
#################################

.PHONY: check_claudesync
# Internal helper: Checks if claudesync is installed locally
check_claudesync:
	@if ! command -v claudesync >/dev/null 2>&1; then \
		echo "claudesync is not installed. Please install from: https://github.com/jahwag/ClaudeSync"; \
		exit 1; \
	fi

.PHONY: claudesync_instructions
claudesync_instructions: ## Show instructions for setting up claudesync
	@echo "PATH Claudesync\n"
	@echo "# install claudesync"
	@echo "claudesync install-completion"
	@echo "claudesync auth login"
	@echo "# Follow instructions"
	@echo "claudesync project create"
	@echo "# check the project was created"
	@echo "claudesync push"

.PHONY: claudesync_setup
claudesync_setup: check_claudesync ## Set up claudesync with all required steps
	@echo "Setting up claudesync..."
	@claudesync install-completion
	@claudesync auth login
	@claudesync project create
	@echo "claudesync setup complete!"

.PHONY: claudesync_push_docs
claudesync_push_docs: check_claudesync ## Push documentation updates to Claude
	@cp .claudeignore_docs .claudeignore
	@claudesync push
	@rm .claudeignore

.PHONY: claudesync_push_code
claudesync_push_code: check_claudesync ## Push code updates to Claude
	@cp .claudeignore_code .claudeignore
	@claudesync push
	@rm .claudeignore
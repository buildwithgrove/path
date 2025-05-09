######################
### Claude targets ###
######################

.PHONY: claudesync_check
# Internal helper: Checks if claudesync is installed locally
claudesync_check:
	@if ! command -v claudesync >/dev/null 2>&1; then \
		echo "claudesync is not installed. Make sure you review this file: docusaurus/docs/develop/tips/claude.md"; \
		exit 1; \
	fi

.PHONY: claudesync_init_docs
claudesync_init_docs: claudesync_check ## Initializes a new ClaudeSync project for documentation
	@echo "###############################################"
	@echo "Initializing a new ClaudeSync project for PATH's documentation"
	@echo "Follow the Remote URL outputted and copy-paste the recommended system prompt"
	@echo "###############################################"
	@claudesync config category add --description markdown_files --patterns "*.md" markdown
	@claudesync project init --new --name path_docs --description "PATH Documentation" --local-path .
	@echo "See this file for the PATH Docs recommended system prompt: docusaurus/docs/develop/tips/claude.md"

.PHONY: claudesync_set_docs
claudesync_set_docs: claudesync_check ## Sets the current ClaudeSync project to the path_docs project
	@echo "Updating the .claudeignore file for documentation"
	@cp .claudeignore_docs .claudeignore
	@echo "Select 'path_docs' from the list of projects"
	@claudesync project set

.PHONY: claudesync_push_docs
claudesync_push_docs: claudesync_check ## Pushes only markdown documentation to Claude
	@echo "Pushing only markdown documentation to Claude..."
	@claudesync push --category markdown
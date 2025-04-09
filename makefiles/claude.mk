#####################
### Claude targets ###
#####################

.PHONY: claudesync_check
# Internal helper: Checks if claudesync is installed locally
claudesync_check:
	@if ! command -v claudesync >/dev/null 2>&1; then \
		echo "claudesync is not installed. Make sure you review this file: docusaurus/docs/develop/tips/claude.md"; \
		exit 1; \
	fi

.PHONY: claudesync_init
claudesync_init: claudesync_check ## Initializes a new ClaudeSync project for documentation
	@echo "###############################################"
	@echo "Initializing a new ClaudeSync project for documentation"
	@echo "When prompted, enter the following name: project_docs"
	@echo "When prompted, enter the following description: Project Documentation"
	@echo "When prompted for an absolute path, press enter"
	@echo "Follow the Remote URL outputted and copy-paste the recommended system prompt"
	@echo "###############################################"
	@claudesync project init --new --name project_docs --description "Project Documentation" --local-path .
	@claudesync config category add --description markdown_files --patterns "*.md" markdown

.PHONY: claudesync_push
claudesync_push: claudesync_check ## Pushes the current project to the ClaudeSync project
	@echo "Pushing all changes to Claude..."
	@claudesync push

.PHONY: claudesync_push_docs
claudesync_push_docs: claudesync_check ## Pushes only markdown documentation to Claude
	@echo "Pushing only markdown documentation to Claude..."
	@claudesync push --category markdown

.PHONY: claudesync_set_docs
claudesync_set_docs: claudesync_check ## Sets the current ClaudeSync project to documentation
	@echo "Updating the .claudeignore file for documentation"
	@cp .claudeignore_docs .claudeignore
	@echo "Select 'project_docs' from the list of projects"
	@claudesync project set

.PHONY: claudesync_categories
claudesync_categories: claudesync_check ## Lists all available categories in the ClaudeSync project
	@claudesync config category ls

.PHONY: claudesync_add_category
claudesync_add_category: claudesync_check ## Adds a new category to the ClaudeSync project
	@echo "Enter category name (e.g., markdown):"
	@read category_name; \
	echo "Enter category description (e.g., markdown_files):"; \
	read category_desc; \
	echo "Enter file patterns (e.g., *.md):"; \
	read patterns; \
	claudesync config category add --description $category_desc --patterns "$patterns" $category_name

.PHONY: claudesync_push_category
claudesync_push_category: claudesync_check ## Pushes only a specific category to Claude
	@echo "Available categories:"
	@claudesync config category ls
	@echo "Enter category name to push:"
	@read category_name; \
	claudesync push --category $category_name
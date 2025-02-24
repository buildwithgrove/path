###########################
#### GitHub PR Helpers ####
###########################

.PHONY: check_diff2html
## Internal helper to check if diff2html is installed
check_diff2html:
	@if ! command -v diff2html &> /dev/null; then \
		echo "########################################################################"; \
		echo "ERROR: diff2html is not installed."; \
		echo "Please install it by running:"; \
		echo "  npm install -g diff2html"; \
		echo "For more information, visit: https://diff2html.xyz/"; \
		echo "########################################################################"; \
		exit 1; \
	fi

.PHONY: git_diff_copy_for_pr_description
git_diff_copy_for_pr_description: check_diff2html ## Copies a Git diff (excluding specific files) to clipboard in JSON format
	@echo "################################################################################"
	@echo "Make sure to track (i.e. git add) all files you want to be included in the diff."
	@echo "################################################################################"
	@git --no-pager diff main \
		-- ':!*.pb.go' ':!*.pulsar.go' ':!*.json' ':!*.yaml' ':!*.yml' | \
		diff2html -s side --format json -i stdin -o stdout | \
		pbcopy
	@echo "Git diff copied to clipboard in JSON format."

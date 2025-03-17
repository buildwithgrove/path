#########################################
### GUARD Initialization Make Targets ###
#########################################

.PHONY: copy_values_yaml
copy_values_yaml: ## Copies the GUARD values template file to the local directory.
	@if [ ! -f ./local/path/.values.yaml ]; then \
		cp ./local/path/values.tmpl.yaml ./local/path/.values.yaml; \
		echo "################################################################"; \
		echo "Created ./local/path/.values.yaml"; \
		echo "################################################################"; \
	else \
		echo "################################################################"; \
		echo "Warning: ./local/path/.values.yaml already exists"; \
		echo "To recreate the file, delete it first and run this command again"; \
		echo "	rm ./local/path/.values.yaml"; \
		echo "	make copy_values_yaml"; \
		echo "################################################################"; \
	fi

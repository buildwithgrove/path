#########################################
### GUARD Initialization Make Targets ###
#########################################

.PHONY: copy_local_config
copy_local_config: copy_path_config copy_guard_config ## Runs copy_path_config and copy_guard_config

.PHONY: copy_path_config
copy_path_config: ## Copies the Shannon PATH configuration file to the local directory.
	@if [ ! -f ./local/path/.config.yaml ]; then \
		cp ./config/examples/config.shannon_example.yaml ./local/path/.config.yaml; \
		echo "################################################################"; \
		echo "Created ./local/path/.config.yaml"; \
		echo ""; \
		echo "Next steps:"; \
		echo ""; \
		echo "For external contributors:"; \
		echo "  Update the following values in .shannon.config.yaml:"; \
		echo "    - gateway_private_key_hex"; \
		echo "    - owned_apps_private_keys_hex"; \
		echo "################################################################"; \
		echo "################################################################"; \
	else \
		echo "################################################################"; \
		echo "Warning: ./local/path/.config.yaml already exists"; \
		echo "To recreate the file, delete it first and run this command again"; \
		echo "	rm ./local/path/.config.yaml"; \
		echo "	make copy_path_config"; \
		echo "################################################################"; \
	fi
	@if [ ! -f ./local/path/.values.yaml ]; then \
		cp ./local/path/values.template.yaml ./local/path/.values.yaml; \
		echo "################################################################"; \
		echo "Created ./local/path/.values.yaml"; \
		echo "################################################################"; \
	else \
		echo "################################################################"; \
		echo "Warning: ./local/path/.values.yaml already exists"; \
		echo "To recreate the file, delete it first and run this command again"; \
		echo "	rm ./local/path/.values.yaml"; \
		echo "	make copy_path_config"; \
		echo "################################################################"; \
	fi

.PHONY: copy_guard_config
copy_guard_config: ## Substitutes the sensitive 0Auth environment variables in the template envoy configuration yaml file and outputs the result to .envoy.yaml
	@if [ ! -f ./local/guard/.values.yaml ]; then \
		cp ./local/guard/values.template.yaml ./local/guard/.values.yaml; \
		echo "################################################################"; \
		echo "Created ./local/guard/.values.yaml"; \
		echo "################################################################"; \
	else \
		echo "################################################################"; \
		echo "Warning: ./local/guard/.values.yaml already exists"; \
		echo "To recreate the file, delete it first and run this command again"; \
		echo "	rm ./local/guard/.values.yaml"; \
		echo "	make copy_guard_config"; \
		echo "################################################################"; \
	fi

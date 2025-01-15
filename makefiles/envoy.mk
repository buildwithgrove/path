#########################################
### Envoy Initialization Make Targets ###
#########################################

.PHONY: init_envoy
init_envoy: copy_envoy_config copy_gateway_endpoints ## Runs copy_envoy_config and copy_gateway_endpoints

.PHONY: copy_envoy_config
copy_envoy_config: ## Substitutes the sensitive 0Auth environment variables in the template envoy configuration yaml file and outputs the result to .envoy.yaml
	@mkdir -p local/path/envoy;
	@./envoy/scripts/copy_envoy_config.sh;

.PHONY: copy_gateway_endpoints
copy_gateway_endpoints: ## Copies the example gateway endpoints YAML file from the PADS repo to ./local/path/envoy/.gateway-endpoints.yaml
	@mkdir -p local/path/envoy;
	@./envoy/scripts/copy_gateway_endpoints_yaml.sh;
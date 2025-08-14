########################
#### Documentation  ####
########################

.PHONY: go_docs
go_docs: ## Start Go documentation server
	@echo "Visit http://localhost:6060/pkg/github.com/buildwithgrove/path"
	godoc -http=:6060

.PHONY: docusaurus_start
docusaurus_start: ## Start docusaurus server
	(cd docusaurus && yarn install && yarn start --port 4000)

.PHONY: docs_serve
docs_serve: docusaurus_start ## Start documentation server (alias for docusaurus_start)

# Uses https://github.com/PaloAltoNetworks/docusaurus-openapi-docs to generate OpenAPI docs.
# This is a custom plugin for Docusaurus that allows us to embed the OpenAPI spec into the docs.
# Outputs docs files to docusaurus/docs/learn/api/*.mdx
.PHONY: docusaurus_gen_api_docs
docusaurus_gen_api_docs: ## Generate docusaurus OpenAPI docs
	(cd docusaurus && yarn install && yarn docusaurus clean-api-docs path && yarn docusaurus gen-api-docs path)

# Outputs docs files to docusaurus/docs/learn/qos/1_supported_services.md
.PHONY: gen_service_qos_docs
gen_service_qos_docs: ## Generate service QoS documentation
	./docusaurus/scripts/service_qos_doc_generator.sh ./config/service_qos_config.go ./docusaurus/docs/learn/qos/1_supported_services.md

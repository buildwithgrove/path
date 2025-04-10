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

.PHONY: gen_service_qos_docs
gen_service_qos_docs: ## Generate service QoS documentation
	./docusaurus/scripts/service_qos.sh ./config/service_qos_config.go ./docusaurus/docs/learn/qos/supported_services.md

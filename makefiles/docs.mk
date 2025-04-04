########################
#### Documentation  ####
########################

.PHONY: go_docs
go_docs: ## Start Go documentation server
	@echo "Visit http://localhost:6060/pkg/github.com/buildwithgrove/path"
	godoc -http=:6060

.PHONY: docusaurus_start
docusaurus_start: ## Start docusaurus server
	(cd docusaurus && yarn install && yarn start)
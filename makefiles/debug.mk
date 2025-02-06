######################
### Debgug targets ###
######################

# TODO_TECHDEBT(@adshmh): Remove host networking mode flag i.e. --network=host, and use a standard method of accessing 
# port 6060 on localhost from within the container.
#
.PHONY: debug_goroutines
# Debugging helper: show goroutines pprof data on port 8081.
# This adds graphviz to a golang docker container to remove the need for installing graphviz,
# since graphviz does not have a general, distribution-independent, installation script:
# https://www.graphviz.org/download
debug_goroutines: check_docker
	@echo "###########################################################################"
	@echo "Starting a golang docker container with graphviz to display goruntime data."
	@echo "###########################################################################"
	@docker run --rm \
		--network=host \
		golang:1.23.6-alpine3.20 \
		apk add --no-cache graphviz && \
		go tool pprof -http="0.0.0.0:8081" http://localhost:6060/debug/pprof/goroutine

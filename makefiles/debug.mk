######################
### Debgug targets ###
######################

.PHONY: debug_goroutines
# Debugging helper: show goroutines pprof data on port 3333
debug_goroutines:
	@go tool pprof -http=:3333 http://localhost:6060/debug/pprof/goroutine

#######################
### Proto  Helpers ####
#######################

.PHONY: proto_gen
proto_gen: proto_gen_observation proto_gen_envoy

.PHONY: proto_gen_observation
proto_gen_observation: ## Generate protobuf artifacts
	protoc -I=./proto --go_out=./observation --go_opt=module='github.com/buildwithgrove/path/observation' \
		./proto/path/*.proto \
		./proto/path/protocol/*.proto \
		./proto/path/qos/*.proto

# TODO_IMPROVE(@commoddity): update to use go:generate comments in the interface files and update this target
.PHONY: proto_gen_envoy
proto_gen_envoy:
	## Generate Go code from the gateway_endpoint.proto file
	protoc --go_out=./envoy/auth_server/proto --go-grpc_out=./envoy/auth_server/proto envoy/auth_server/proto/gateway_endpoint.proto

.PHONY: proto_clean
proto_clean: ## Delete existing .pb.go files
	find . -name "*.pb.go" -delete

.PHONY: proto_regen
proto_regen: proto_clean proto_gen ## Regenerate protobuf artifacts

#######################
### Proto  Helpers ####
#######################

.PHONY: proto_gen
proto_gen: proto_gen_observation proto_gen_envoy ## Generate all protobuf artifacts

.PHONY: proto_regen
proto_regen: proto_clean proto_gen ## Regenerate all protobuf artifacts

.PHONY: proto_gen_observation
proto_gen_observation: ## Generate observation protobuf artifacts
	protoc -I=./proto \
		--go_out=./observation \
		--go_opt=module='github.com/buildwithgrove/path/observation' \
		./proto/path/*.proto \
		./proto/path/protocol/*.proto \
		./proto/path/qos/*.proto

# TODO_IMPROVE(@commoddity): update to use go:generate comments in the interface files and update this target
.PHONY: proto_gen_envoy
proto_gen_envoy: ## Generate envoy protobuf artifacts
	protoc \
		--go_out=./envoy/auth_server/proto \
		--go-grpc_out=./envoy/auth_server/proto \
		envoy/auth_server/proto/gateway_endpoint.proto

.PHONY: proto_clean
proto_clean: ## Delete existing protobuf artifacts (i.e. .pb.go files)
	find . -name "*.pb.go" -delete

.PHONY: proto_mock_gen
proto_mock_gen: ## Generate mocks for protobuf artifacts
	mockgen -source=./envoy/auth_server/auth/auth_handler.go -destination=./envoy/auth_server/auth/endpoint_store_mock_test.go -package=auth
	mockgen -source=./router/router.go -destination=./router/router_mock_test.go -package=router
	mockgen -source=./envoy/auth_server/proto/gateway_endpoint_grpc.pb.go -destination=./envoy/auth_server/endpoint_store/client_mock_test.go -package=endpointstore -mock_names=GatewayEndpointsClient=MockGatewayEndpointsClient

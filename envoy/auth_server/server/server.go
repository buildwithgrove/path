//go:build auth_server

package server

import (
	"context"
	"fmt"

	auth_pb "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/pokt-network/poktroll/pkg/polylog"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/buildwithgrove/auth-server/user"
)

// The userDataCache contains an in-memory cache of GatewayEndpoints
// and their associated data from the connected Postgres database.
type (
	userDataCache interface {
		GetGatewayEndpoint(user.EndpointID) (user.GatewayEndpoint, bool)
	}
	jwtParser interface {
		ParseJWT(req *auth_pb.CheckRequest) (user.EndpointID, *errorResponse)
	}
)

// struct with check method
type AuthServer struct {
	JWTParser jwtParser
	Cache     userDataCache
	Logger    polylog.Logger
}

func (a *AuthServer) Check(ctx context.Context, req *auth_pb.CheckRequest,
) (*auth_pb.CheckResponse, error) {
	fmt.Println("All request goes through me")

	path := req.GetAttributes().GetRequest().GetHttp().GetPath()
	if path == "" {
		return &auth_pb.CheckResponse{
			Status: &status.Status{
				Code:    int32(codes.InvalidArgument),
				Message: "Path not provided",
			},
		}, nil
	}

	// If the path is "/healthz", we don't need to authenticate
	if path == "/healthz" {
		return &auth_pb.CheckResponse{
			Status: &status.Status{
				Code:    int32(codes.OK),
				Message: "OK",
			},
		}, nil
	}

	// Parse the JWT and extract the endpoint ID
	endpointID, errResp := a.JWTParser.ParseJWT(req)
	if errResp != nil {
		return &auth_pb.CheckResponse{
			Status: &status.Status{
				Code:    int32(codes.NotFound),
				Message: errResp.message,
			},
		}, nil
	}

	// If GatewayEndpoint is not found send an error response downstream (client)
	gatewayEndpoint, ok := a.getGatewayEndpoint(endpointID)
	if !ok {
		return &auth_pb.CheckResponse{
			Status: &status.Status{
				Code:    int32(codes.NotFound),
				Message: "Endpoint not found",
			},
		}, nil
	}

	// Add endpoint ID and rate limiting values to the dynamic metadata
	// to be passed to the next service in the chain (rate limiting)
	dynamicMetadata := a.getDynamicMetadata(gatewayEndpoint)

	// Return a valid response
	return &auth_pb.CheckResponse{
		Status: &status.Status{
			Code:    int32(codes.OK),
			Message: "OK",
		},
		DynamicMetadata: dynamicMetadata,
	}, nil
}

/* --------------------------------- Service Request Processing -------------------------------- */

const (
	reqHeaderEndpointID          = "x-endpoint-id"    // Set on all service requests
	reqHeaderRateLimitEndpointID = "x-rl-endpoint-id" // Set only on service requests that should be rate limited
	reqHeaderRateLimitThroughput = "x-rl-throughput"  // Set only on service requests that should be rate limited
)

// getGatewayEndpoint fetches the GatewayEndpoint from the database and a bool indicating if it was found
func (a *AuthServer) getGatewayEndpoint(endpointID user.EndpointID) (user.GatewayEndpoint, bool) {
	return a.Cache.GetGatewayEndpoint(endpointID)
}

// getDynamicMetadata sets all dynamic metadata required by the PATH services on the request being forwarded
func (a *AuthServer) getDynamicMetadata(gatewayEndpoint user.GatewayEndpoint) *structpb.Struct {
	dynamicMetadata := &structpb.Struct{}

	// Set endpoint ID in the dynamic metadata
	dynamicMetadata.Fields[reqHeaderEndpointID] = &structpb.Value{
		Kind: &structpb.Value_StringValue{
			StringValue: string(gatewayEndpoint.EndpointID),
		},
	}

	// Set rate limit headers if the gateway endpoint should be rate limited
	if gatewayEndpoint.RateLimiting.ThroughputLimit > 0 {
		dynamicMetadata.Fields[reqHeaderRateLimitEndpointID] = &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: string(gatewayEndpoint.EndpointID),
			},
		}
		dynamicMetadata.Fields[reqHeaderRateLimitThroughput] = &structpb.Value{
			Kind: &structpb.Value_NumberValue{
				NumberValue: float64(gatewayEndpoint.RateLimiting.ThroughputLimit),
			},
		}
	}

	return dynamicMetadata
}

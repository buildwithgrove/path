package endpointstore

import (
	"context"
	"io"
	"testing"

	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	protoPkg "google.golang.org/protobuf/proto"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

// MockStream is a mock implementation of the grpc.ClientStream interface.
type MockStream struct {
	grpc.ClientStream
	updates chan *proto.AuthDataUpdate
}

func (m *MockStream) Recv() (*proto.AuthDataUpdate, error) {
	update := <-m.updates
	if update == nil {
		return nil, io.EOF
	}
	return update, nil
}

func newTestStore(t *testing.T, ctx context.Context, updates chan *proto.AuthDataUpdate, ctrl *gomock.Controller) *endpointStore {
	mockClient := NewMockGatewayEndpointsClient(ctrl)

	// Set up the expected call for FetchAuthDataSync
	mockClient.EXPECT().FetchAuthDataSync(gomock.Any(), gomock.Any()).Return(getTestGatewayEndpoints(), nil)

	// Set up the expected call for StreamUpdates
	mockStream := &MockStream{updates: updates}
	mockClient.EXPECT().StreamAuthDataUpdates(gomock.Any(), gomock.Any()).Return(mockStream, nil).AnyTimes()

	store, err := NewEndpointStore(ctx, polyzero.NewLogger(), mockClient)
	require.NoError(t, err)

	return store
}

func Test_GetGatewayEndpoint(t *testing.T) {
	tests := []struct {
		name                    string
		endpointID              string
		expectedGatewayEndpoint *proto.GatewayEndpoint
		expectedEndpointFound   bool
		update                  *proto.AuthDataUpdate
	}{
		{
			name:                    "should return gateway endpoint when found",
			endpointID:              "endpoint_1_static_key",
			expectedGatewayEndpoint: getTestGatewayEndpoints().Endpoints["endpoint_1_static_key"],
			expectedEndpointFound:   true,
		},
		{
			name:                    "should return different gateway endpoint when found",
			endpointID:              "endpoint_2_jwt",
			expectedGatewayEndpoint: getTestGatewayEndpoints().Endpoints["endpoint_2_jwt"],
			expectedEndpointFound:   true,
		},
		{
			name:                    "should return brand new gateway endpoint when update is received for new endpoint",
			endpointID:              "endpoint_3_no_auth",
			update:                  getTestUpdate("endpoint_3_no_auth"),
			expectedGatewayEndpoint: getTestUpdate("endpoint_3_no_auth").GatewayEndpoint,
			expectedEndpointFound:   true,
		},
		{
			name:                    "should return updated existing gateway endpoint when update is received for existing endpoint",
			endpointID:              "endpoint_2_jwt",
			update:                  getTestUpdate("endpoint_2_jwt"),
			expectedGatewayEndpoint: getTestUpdate("endpoint_2_jwt").GatewayEndpoint,
			expectedEndpointFound:   true,
		},
		{
			name:                    "should not return gateway endpoint when update is received to delete endpoint",
			endpointID:              "endpoint_1_static_key",
			update:                  getTestUpdate("endpoint_1_static_key"),
			expectedGatewayEndpoint: nil,
			expectedEndpointFound:   false,
		},
		{
			name:                    "should return false when gateway endpoint not found",
			endpointID:              "endpoint_3_no_auth",
			expectedGatewayEndpoint: nil,
			expectedEndpointFound:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			updates := make(chan *proto.AuthDataUpdate)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			store := newTestStore(t, ctx, updates, ctrl)

			// Send updates for this test case
			if test.update != nil {
				updates <- test.update
			}
			updates <- nil // Signal end of updates

			gatewayEndpoint, found := store.GetGatewayEndpoint(test.endpointID)
			c.Equal(test.expectedEndpointFound, found)
			c.True(protoPkg.Equal(test.expectedGatewayEndpoint, gatewayEndpoint), "expected and actual GatewayEndpoint do not match")
		})
	}
}

// getTestGatewayEndpoints returns a mock response for the initial endpoint store data, received when the endpoint store is first created
func getTestGatewayEndpoints() *proto.AuthDataResponse {
	return &proto.AuthDataResponse{
		Endpoints: map[string]*proto.GatewayEndpoint{
			"endpoint_1_static_key": {
				EndpointId: "endpoint_1_static_key",
				Auth: &proto.Auth{
					AuthType: &proto.Auth_StaticApiKey{
						StaticApiKey: &proto.StaticAPIKey{
							ApiKey: "api_key_1",
						},
					},
				},
				RateLimiting: &proto.RateLimiting{},
				Metadata: &proto.Metadata{
					AccountId: "account_1",
					PlanType:  "PLAN_UNLIMITED",
					Email:     "amos.burton@opa.belt",
				},
			},
			"endpoint_2_jwt": {
				EndpointId: "endpoint_2_jwt",
				Auth: &proto.Auth{
					AuthType: &proto.Auth_Jwt{
						Jwt: &proto.JWT{
							AuthorizedUsers: map[string]*proto.Empty{
								"auth0|user_1": {},
								"auth0|user_2": {},
							},
						},
					},
				},
				RateLimiting: &proto.RateLimiting{},
				Metadata: &proto.Metadata{
					AccountId: "account_2",
					PlanType:  "PLAN_UNLIMITED",
					Email:     "paul.atreides@arrakis.com",
				},
			},
		},
	}
}

// getTestUpdate returns a mock update for a given endpoint ID, used to test the endpoint store's behavior when updates are received
// Will be one of three cases:
// 1. A new GatewayEndpoint was created (endpoint_3)
// 2. An existing GatewayEndpoint was updated (endpoint_2)
// 3. An existing GatewayEndpoint was deleted (endpoint_1)
func getTestUpdate(endpointID string) *proto.AuthDataUpdate {
	updatesMap := map[string]*proto.AuthDataUpdate{
		"endpoint_2_jwt": {
			EndpointId: "endpoint_2_jwt",
			GatewayEndpoint: &proto.GatewayEndpoint{
				EndpointId: "endpoint_2_jwt",
				Auth: &proto.Auth{
					AuthType: &proto.Auth_Jwt{
						Jwt: &proto.JWT{
							AuthorizedUsers: map[string]*proto.Empty{
								"auth0|user_1": {},
								"auth0|user_2": {},
							},
						},
					},
				},
				RateLimiting: &proto.RateLimiting{},
				Metadata: &proto.Metadata{
					AccountId: "account_2",
					PlanType:  "PLAN_UNLIMITED",
					Email:     "paul.atreides@arrakis.com",
				},
			},
			Delete: false,
		},
		"endpoint_3_no_auth": {
			EndpointId: "endpoint_3_no_auth",
			GatewayEndpoint: &proto.GatewayEndpoint{
				EndpointId: "endpoint_3_no_auth",
				Auth: &proto.Auth{
					AuthType: &proto.Auth_NoAuth{},
				},
				RateLimiting: &proto.RateLimiting{
					ThroughputLimit:     30,
					CapacityLimit:       100_000,
					CapacityLimitPeriod: proto.CapacityLimitPeriod_CAPACITY_LIMIT_PERIOD_MONTHLY,
				},
				Metadata: &proto.Metadata{
					AccountId: "account_2",
					PlanType:  "PLAN_FREE",
					Email:     "frodo.baggins@shire.io",
				},
			},
			Delete: false,
		},
		"endpoint_1_static_key": {
			EndpointId: "endpoint_1_static_key",
			Delete:     true,
		},
	}

	return updatesMap[endpointID]
}

//go:build auth_plugin

package db

import (
	"context"
	"testing"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/buildwithgrove/authorizer-plugin/types"
)

func Test_GetGatewayEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		endpointID types.EndpointID
		mockReturn map[types.EndpointID]types.GatewayEndpoint
		expected   types.GatewayEndpoint
		found      bool
	}{
		{
			name:       "should return gateway endpoint when found",
			endpointID: "endpoint_1",
			mockReturn: getTestGatewayEndpoints(),
			expected:   getTestGatewayEndpoints()["endpoint_1"],
			found:      true,
		},
		{
			name:       "should return different gateway endpoint when found",
			endpointID: "endpoint_2",
			mockReturn: getTestGatewayEndpoints(),
			expected:   getTestGatewayEndpoints()["endpoint_2"],
			found:      true,
		},
		{
			name:       "should return false when gateway endpoint not found",
			endpointID: "endpoint_3",
			mockReturn: getTestGatewayEndpoints(),
			expected:   types.GatewayEndpoint{},
			found:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			ctrl := gomock.NewController(t)

			mockDB := NewMockDBDriver(ctrl)
			mockDB.EXPECT().GetGatewayEndpoints(gomock.Any()).Return(test.mockReturn, nil)

			cache, err := NewUserDataCache(mockDB, time.Minute, polyzero.NewLogger())
			c.NoError(err)

			gatewayEndpoint, found := cache.GetGatewayEndpoint(context.Background(), test.endpointID)
			c.Equal(test.found, found)
			c.Equal(test.expected, gatewayEndpoint)
		})
	}
}

func Test_cacheRefreshHandler(t *testing.T) {
	tests := []struct {
		name       string
		mockReturn map[types.EndpointID]types.GatewayEndpoint
		expected   map[types.EndpointID]types.GatewayEndpoint
	}{
		{
			name:       "should refresh cache with new data",
			mockReturn: map[types.EndpointID]types.GatewayEndpoint{"endpoint_1": {EndpointID: "endpoint_1"}},
			expected:   map[types.EndpointID]types.GatewayEndpoint{"endpoint_1": {EndpointID: "endpoint_1"}},
		},
		{
			name:       "should handle empty cache refresh",
			mockReturn: map[types.EndpointID]types.GatewayEndpoint{},
			expected:   map[types.EndpointID]types.GatewayEndpoint{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			ctrl := gomock.NewController(t)

			mockDB := NewMockDBDriver(ctrl)
			mockDB.EXPECT().GetGatewayEndpoints(gomock.Any()).Return(test.mockReturn, nil).AnyTimes()

			cache, err := NewUserDataCache(mockDB, time.Minute, polyzero.NewLogger())
			c.NoError(err)

			cache.cacheRefreshInterval = time.Millisecond * 10

			go cache.cacheRefreshHandler(context.Background())

			time.Sleep(time.Millisecond * 20)

			c.Equal(test.expected, cache.gatewayEndpoints)
		})
	}
}

func Test_updateCache(t *testing.T) {
	tests := []struct {
		name       string
		mockReturn map[types.EndpointID]types.GatewayEndpoint
		expected   map[types.EndpointID]types.GatewayEndpoint
	}{
		{
			name:       "should update cache with gateway endpoints",
			mockReturn: map[types.EndpointID]types.GatewayEndpoint{"endpoint_1": {EndpointID: "endpoint_1"}},
			expected:   map[types.EndpointID]types.GatewayEndpoint{"endpoint_1": {EndpointID: "endpoint_1"}},
		},
		{
			name:       "should handle empty gateway endpoints",
			mockReturn: map[types.EndpointID]types.GatewayEndpoint{},
			expected:   map[types.EndpointID]types.GatewayEndpoint{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			ctrl := gomock.NewController(t)

			mockDB := NewMockDBDriver(ctrl)
			mockDB.EXPECT().GetGatewayEndpoints(gomock.Any()).Return(test.mockReturn, nil).AnyTimes()

			cache, err := NewUserDataCache(mockDB, time.Minute, polyzero.NewLogger())
			c.NoError(err)

			err = cache.updateCache(context.Background())
			c.NoError(err)
			c.Equal(test.expected, cache.gatewayEndpoints)
		})
	}
}

func getTestGatewayEndpoints() map[types.EndpointID]types.GatewayEndpoint {
	return map[types.EndpointID]types.GatewayEndpoint{
		"endpoint_1": {
			EndpointID: "endpoint_1",
			Auth: types.Auth{
				APIKey:         "api_key_1",
				APIKeyRequired: true,
			},
			UserAccount: types.UserAccount{
				AccountID: "account_1",
				PlanType:  "PLAN_FREE",
			},
			RateLimiting: types.RateLimiting{
				ThroughputLimit: 30,
				CapacityLimit:   100,
			},
		},
		"endpoint_2": {
			EndpointID: "endpoint_2",
			Auth: types.Auth{
				APIKey:         "api_key_2",
				APIKeyRequired: true,
			},
			UserAccount: types.UserAccount{
				AccountID: "account_2",
				PlanType:  "PLAN_UNLIMITED",
			},
			RateLimiting: types.RateLimiting{
				ThroughputLimit: 50,
				CapacityLimit:   200,
			},
		},
	}
}

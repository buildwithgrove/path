package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/buildwithgrove/auth-server/user"
)

var connectionString string

func TestMain(m *testing.M) {
	// Initialize the ephemeral postgres docker container
	pool, resource, databaseURL := setupPostgresDocker()
	connectionString = databaseURL

	// Run DB integration test
	exitCode := m.Run()

	// Cleanup the ephemeral postgres docker container
	cleanupPostgresDocker(m, pool, resource)
	os.Exit(exitCode)
}

func Test_Integration_GetGatewayEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		expected map[user.EndpointID]user.GatewayEndpoint
	}{
		{
			name: "should retrieve all gateway endpoints correctly",
			expected: map[user.EndpointID]user.GatewayEndpoint{
				"endpoint_1": {
					EndpointID: "endpoint_1",
					Auth: user.Auth{
						AuthorizedUsers: map[user.ProviderUserID]struct{}{
							"auth0|user_1": {},
							"auth0|user_4": {},
						},
					},
					UserAccount: user.UserAccount{
						AccountID: "account_1",
						PlanType:  "PLAN_FREE",
					},
					RateLimiting: user.RateLimiting{
						ThroughputLimit:     30,
						CapacityLimit:       100000,
						CapacityLimitPeriod: "daily",
					},
				},
				"endpoint_2": {
					EndpointID: "endpoint_2",
					Auth: user.Auth{
						AuthorizedUsers: map[user.ProviderUserID]struct{}{
							"auth0|user_2": {},
						},
					},
					UserAccount: user.UserAccount{
						AccountID: "account_2",
						PlanType:  "PLAN_UNLIMITED",
					},
					RateLimiting: user.RateLimiting{
						ThroughputLimit: 0,
						CapacityLimit:   0,
					},
				},
				"endpoint_3": {
					EndpointID: "endpoint_3",
					Auth: user.Auth{
						AuthorizedUsers: map[user.ProviderUserID]struct{}{
							"auth0|user_3": {},
						},
					},
					UserAccount: user.UserAccount{
						AccountID: "account_3",
						PlanType:  "PLAN_FREE",
					},
					RateLimiting: user.RateLimiting{
						ThroughputLimit:     30,
						CapacityLimit:       100000,
						CapacityLimitPeriod: "daily",
					},
				},
				"endpoint_4": {
					EndpointID: "endpoint_4",
					Auth: user.Auth{
						AuthorizedUsers: map[user.ProviderUserID]struct{}{
							"auth0|user_1": {},
							"auth0|user_4": {},
						},
					},
					UserAccount: user.UserAccount{
						AccountID: "account_1",
						PlanType:  "PLAN_FREE",
					},
					RateLimiting: user.RateLimiting{
						ThroughputLimit:     30,
						CapacityLimit:       100000,
						CapacityLimitPeriod: "daily",
					},
				},
				"endpoint_5": {
					EndpointID: "endpoint_5",
					Auth: user.Auth{
						AuthorizedUsers: map[user.ProviderUserID]struct{}{
							"auth0|user_2": {},
						},
					},
					UserAccount: user.UserAccount{
						AccountID: "account_2",
						PlanType:  "PLAN_UNLIMITED",
					},
					RateLimiting: user.RateLimiting{
						ThroughputLimit: 0,
						CapacityLimit:   0,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if testing.Short() {
				t.Skip("skipping driver integration test")
			}

			c := require.New(t)

			driver, cleanup, err := NewPostgresDriver(connectionString)
			c.NoError(err)
			defer func() {
				_ = cleanup()
			}()

			endpoints, err := driver.GetGatewayEndpoints(context.Background())
			c.NoError(err)
			c.Equal(test.expected, endpoints)
		})
	}
}

func Test_convertToGatewayEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		rows     []SelectGatewayEndpointsRow
		expected map[user.EndpointID]user.GatewayEndpoint
		wantErr  bool
	}{
		{
			name: "should convert rows to gateway endpoints successfully",
			rows: []SelectGatewayEndpointsRow{
				{
					ID:                      "endpoint_1",
					AccountID:               pgtype.Text{String: "account_1", Valid: true},
					Plan:                    pgtype.Text{String: "PLAN_FREE", Valid: true},
					RateLimitThroughput:     pgtype.Int4{Int32: 30, Valid: true},
					RateLimitCapacity:       pgtype.Int4{Int32: 100000, Valid: true},
					RateLimitCapacityPeriod: NullRateLimitCapacityPeriod{RateLimitCapacityPeriod: "daily", Valid: true},
					ProviderUserIds:         []string{"auth0|user_1", "auth0|user_4"},
				},
			},
			expected: map[user.EndpointID]user.GatewayEndpoint{
				"endpoint_1": {
					EndpointID: "endpoint_1",
					Auth: user.Auth{
						AuthorizedUsers: map[user.ProviderUserID]struct{}{
							"auth0|user_1": {},
							"auth0|user_4": {},
						},
					},
					UserAccount: user.UserAccount{
						AccountID: "account_1",
						PlanType:  "PLAN_FREE",
					},
					RateLimiting: user.RateLimiting{
						ThroughputLimit:     30,
						CapacityLimit:       100000,
						CapacityLimitPeriod: "daily",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			driver := &postgresDriver{}
			endpoints, err := driver.convertToGatewayEndpoints(test.rows)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expected, endpoints)
			}
		})
	}
}

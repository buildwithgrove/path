package db

import (
	"context"
	"testing"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/buildwithgrove/path/user"
)

func Test_GetUserApp(t *testing.T) {
	tests := []struct {
		name       string
		userAppID  user.UserAppID
		mockReturn map[user.UserAppID]user.UserApp
		expected   user.UserApp
		found      bool
	}{
		{
			name:       "should return user app when found",
			userAppID:  "user_app_1",
			mockReturn: getTestUserApps(),
			expected:   getTestUserApps()["user_app_1"],
			found:      true,
		},
		{
			name:       "should return different user app when found",
			userAppID:  "user_app_2",
			mockReturn: getTestUserApps(),
			expected:   getTestUserApps()["user_app_2"],
			found:      true,
		},
		{
			name:       "should return false when user app not found",
			userAppID:  "user_app_3",
			mockReturn: getTestUserApps(),
			expected:   user.UserApp{},
			found:      false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			ctrl := gomock.NewController(t)

			mockDB := NewMockDriver(ctrl)
			mockDB.EXPECT().GetUserApps(gomock.Any()).Return(test.mockReturn, nil)

			cache, err := NewCache(mockDB, time.Minute, polyzero.NewLogger())
			c.NoError(err)

			userApp, found := cache.GetUserApp(context.Background(), test.userAppID)
			c.Equal(test.found, found)
			c.Equal(test.expected, userApp)
		})
	}
}

func Test_cacheRefreshHandler(t *testing.T) {
	tests := []struct {
		name       string
		mockReturn map[user.UserAppID]user.UserApp
		expected   map[user.UserAppID]user.UserApp
	}{
		{
			name:       "should refresh cache with new data",
			mockReturn: map[user.UserAppID]user.UserApp{"user_app_1": {ID: "user_app_1"}},
			expected:   map[user.UserAppID]user.UserApp{"user_app_1": {ID: "user_app_1"}},
		},
		{
			name:       "should handle empty cache refresh",
			mockReturn: map[user.UserAppID]user.UserApp{},
			expected:   map[user.UserAppID]user.UserApp{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			ctrl := gomock.NewController(t)

			mockDB := NewMockDriver(ctrl)
			mockDB.EXPECT().GetUserApps(gomock.Any()).Return(test.mockReturn, nil).AnyTimes()

			cache, err := NewCache(mockDB, time.Minute, polyzero.NewLogger())
			c.NoError(err)

			cache.cacheRefreshInterval = time.Millisecond * 10

			go cache.cacheRefreshHandler(context.Background())

			time.Sleep(time.Millisecond * 20)

			c.Equal(test.expected, cache.userApps)
		})
	}
}

func Test_setCache(t *testing.T) {
	tests := []struct {
		name       string
		mockReturn map[user.UserAppID]user.UserApp
		expected   map[user.UserAppID]user.UserApp
	}{
		{
			name:       "should set cache with user apps",
			mockReturn: map[user.UserAppID]user.UserApp{"user_app_1": {ID: "user_app_1"}},
			expected:   map[user.UserAppID]user.UserApp{"user_app_1": {ID: "user_app_1"}},
		},
		{
			name:       "should handle empty user apps",
			mockReturn: map[user.UserAppID]user.UserApp{},
			expected:   map[user.UserAppID]user.UserApp{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			ctrl := gomock.NewController(t)

			mockDB := NewMockDriver(ctrl)
			mockDB.EXPECT().GetUserApps(gomock.Any()).Return(test.mockReturn, nil).AnyTimes()

			cache, err := NewCache(mockDB, time.Minute, polyzero.NewLogger())
			c.NoError(err)

			err = cache.setCache(context.Background())
			c.NoError(err)
			c.Equal(test.expected, cache.userApps)
		})
	}
}

func getTestUserApps() map[user.UserAppID]user.UserApp {
	return map[user.UserAppID]user.UserApp{
		"user_app_1": {
			ID:                  "user_app_1",
			AccountID:           "account_1",
			PlanType:            "PLAN_FREE",
			SecretKey:           "secret_1",
			SecretKeyRequired:   true,
			RateLimitThroughput: 30,
			Allowlists: map[user.AllowlistType]map[string]struct{}{
				user.AllowlistTypeOrigins:   {"origin_1": {}},
				user.AllowlistTypeContracts: {"contract_1": {}},
			},
		},
		"user_app_2": {
			ID:                "user_app_2",
			AccountID:         "account_2",
			PlanType:          "PLAN_UNLIMITED",
			SecretKey:         "secret_2",
			SecretKeyRequired: true,
			Allowlists: map[user.AllowlistType]map[string]struct{}{
				user.AllowlistTypeOrigins:   {"origin_2": {}},
				user.AllowlistTypeContracts: {"contract_2": {}},
			},
		},
	}
}

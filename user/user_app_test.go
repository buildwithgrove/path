package user

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_IsContractAllowed(t *testing.T) {
	tests := []struct {
		name       string
		app        UserApp
		contractID string
		expected   bool
	}{
		{
			name: "should return true if contract is allowlisted",
			app: UserApp{
				Allowlists: map[AllowlistType]map[string]struct{}{
					AllowlistTypeContracts: {"contract_1": {}},
				},
			},
			contractID: "contract_1",
			expected:   true,
		},
		{
			name: "should return false if contract is not allowlisted",
			app: UserApp{
				Allowlists: map[AllowlistType]map[string]struct{}{
					AllowlistTypeContracts: {},
				},
			},
			contractID: "contract_2",
			expected:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			result := test.app.IsContractAllowed(test.contractID)
			c.Equal(test.expected, result)
		})
	}
}

func Test_IsMethodAllowed(t *testing.T) {
	tests := []struct {
		name     string
		app      UserApp
		method   string
		expected bool
	}{
		{
			name: "should return true if method is allowlisted",
			app: UserApp{
				Allowlists: map[AllowlistType]map[string]struct{}{
					AllowlistTypeMethods: {"method_1": {}},
				},
			},
			method:   "method_1",
			expected: true,
		},
		{
			name: "should return false if method is not allowlisted",
			app: UserApp{
				Allowlists: map[AllowlistType]map[string]struct{}{
					AllowlistTypeMethods: {},
				},
			},
			method:   "method_2",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			result := test.app.IsMethodAllowed(test.method)
			c.Equal(test.expected, result)
		})
	}
}

func Test_IsOriginAllowed(t *testing.T) {
	tests := []struct {
		name     string
		app      UserApp
		origin   string
		expected bool
	}{
		{
			name: "should return true if origin is allowlisted",
			app: UserApp{
				Allowlists: map[AllowlistType]map[string]struct{}{
					AllowlistTypeOrigins: {"origin_1": {}},
				},
			},
			origin:   "origin_1",
			expected: true,
		},
		{
			name: "should return false if origin is not allowlisted",
			app: UserApp{
				Allowlists: map[AllowlistType]map[string]struct{}{
					AllowlistTypeOrigins: {},
				},
			},
			origin:   "origin_2",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			result := test.app.IsOriginAllowed(test.origin)
			c.Equal(test.expected, result)
		})
	}
}

func Test_IsServiceAllowed(t *testing.T) {
	tests := []struct {
		name     string
		app      UserApp
		service  string
		expected bool
	}{
		{
			name: "should return true if service is allowlisted",
			app: UserApp{
				Allowlists: map[AllowlistType]map[string]struct{}{
					AllowlistTypeServices: {"service_1": {}},
				},
			},
			service:  "service_1",
			expected: true,
		},
		{
			name: "should return false if service is not allowlisted",
			app: UserApp{
				Allowlists: map[AllowlistType]map[string]struct{}{
					AllowlistTypeServices: {},
				},
			},
			service:  "service_2",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			result := test.app.IsServiceAllowed(test.service)
			c.Equal(test.expected, result)
		})
	}
}

func Test_IsUserAgentAllowed(t *testing.T) {
	tests := []struct {
		name      string
		app       UserApp
		userAgent string
		expected  bool
	}{
		{
			name: "should return true if user agent is allowlisted",
			app: UserApp{
				Allowlists: map[AllowlistType]map[string]struct{}{
					AllowlistTypeUserAgents: {"user_agent_1": {}},
				},
			},
			userAgent: "user_agent_1",
			expected:  true,
		},
		{
			name: "should return false if user agent is not allowlisted",
			app: UserApp{
				Allowlists: map[AllowlistType]map[string]struct{}{
					AllowlistTypeUserAgents: {},
				},
			},
			userAgent: "user_agent_2",
			expected:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			result := test.app.IsUserAgentAllowed(test.userAgent)
			c.Equal(test.expected, result)
		})
	}
}

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

package driver

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_IsWhitelisted(t *testing.T) {
	tests := []struct {
		name           string
		app            UserApp
		whitelistType  WhitelistType
		whitelistValue WhitelistValue
		expected       bool
	}{
		{
			name: "should return true if origin is whitelisted",
			app: UserApp{
				Whitelists: map[WhitelistType]map[WhitelistValue]struct{}{
					"origins": {"origin1": {}},
				},
			},
			whitelistType:  "origins",
			whitelistValue: "origin1",
			expected:       true,
		},
		{
			name: "should return false if origin is not whitelisted",
			app: UserApp{
				Whitelists: map[WhitelistType]map[WhitelistValue]struct{}{
					"origins": {},
				},
			},
			whitelistType:  "origins",
			whitelistValue: "origin2",
			expected:       false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			result := test.app.IsWhitelisted(test.whitelistType, test.whitelistValue)
			c.Equal(test.expected, result)
		})
	}
}

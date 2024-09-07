package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserDataConfig_validate(t *testing.T) {
	tests := []struct {
		name    string
		config  UserDataConfig
		wantErr bool
	}{
		{
			name:    "should validate correct DB connection string",
			config:  UserDataConfig{DBConnectionString: "postgres://user:password@localhost:5432/database"},
			wantErr: false,
		},
		{
			name:    "should fail for incorrect DB connection string",
			config:  UserDataConfig{DBConnectionString: "invalid_connection_string"},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			err := test.config.validate()
			if test.wantErr {
				c.Error(err)
			} else {
				c.NoError(err)
			}
		})
	}
}

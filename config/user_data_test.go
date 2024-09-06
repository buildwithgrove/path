package config

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/require"
)

func TestUserDataConfig_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		want     UserDataConfig
		wantErr  bool
	}{
		{
			name: "should unmarshal and validate correct DB connection string",
			yamlData: `
db_connection_string: "postgres://user:password@localhost:5432/database"
`,
			want: UserDataConfig{
				DBConnectionString: "postgres://user:password@localhost:5432/database",
			},
			wantErr: false,
		},
		{
			name: "should return error for invalid DB connection string",
			yamlData: `
db_connection_string: "invalid_connection_string"
`,
			want:    UserDataConfig{},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			var got UserDataConfig
			err := yaml.Unmarshal([]byte(test.yamlData), &got)
			if test.wantErr {
				c.Error(err)
			} else {
				c.NoError(err)
				c.Equal(test.want, got)
			}
		})
	}
}

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

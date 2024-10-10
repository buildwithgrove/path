//go:build auth_server

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_LoadAuthServerConfigFromYAML(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     AuthServerConfig
		wantErr  bool
	}{
		{
			name:     "should load valid auth plugin config without error",
			filePath: "./testdata/plugin.example.yaml",
			want: AuthServerConfig{
				PostgresConnectionString: "postgres://user:password@localhost:5432/database",
				CacheRefreshInterval:     5 * time.Minute,
				JWTIssuer:                "issuer",
				JWTAudience:              "audience",
				JWTJWKSURL:               "https://www.googleapis.com/oauth2/v3/certs",
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			got, err := LoadAuthServerConfigFromYAML(test.filePath)
			if test.wantErr {
				c.Error(err)
			} else {
				c.NoError(err)
				c.Equal(test.want, got)
			}
		})
	}
}

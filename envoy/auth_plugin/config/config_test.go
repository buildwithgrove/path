//go:build auth_plugin

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_LoadAuthorizerPluginConfigFromYAML(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     AuthorizerPluginConfig
		wantErr  bool
	}{
		{
			name:     "should load valid authorizer plugin config without error",
			filePath: "./testdata/plugin.example.yaml",
			want: AuthorizerPluginConfig{
				PostgresConnectionString: "postgres://user:password@localhost:5432/database",
				CacheRefreshInterval:     5 * time.Minute,
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			got, err := LoadAuthorizerPluginConfigFromYAML(test.filePath)
			if test.wantErr {
				c.Error(err)
			} else {
				c.NoError(err)
				c.Equal(test.want, got)
			}
		})
	}
}

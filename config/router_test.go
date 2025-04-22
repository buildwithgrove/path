package config

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRouterConfig_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		want     RouterConfig
		wantErr  bool
	}{
		{
			name: "should unmarshal without error",
			yamlData: `
port: 8080
`,
			want: RouterConfig{
				Port: 8080,
			},
			wantErr: false,
		},
		{
			name: "should return error for invalid YAML",
			yamlData: `
port: invalid_port
`,
			want:    RouterConfig{},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			var got RouterConfig
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

func TestRouterConfig_hydrateRouterDefaults(t *testing.T) {
	tests := []struct {
		name string
		cfg  RouterConfig
		want RouterConfig
	}{
		{
			name: "should set all defaults",
			cfg:  RouterConfig{},
			want: RouterConfig{
				Port:               defaultPort,
				MaxRequestBodySize: defaultMaxRequestBodySize,
				ReadTimeout:        defaultHTTPServerReadTimeout,
				WriteTimeout:       defaultHTTPServerWriteTimeout,
				IdleTimeout:        defaultHTTPServerIdleTimeout,
			},
		},
		{
			name: "should not override set values",
			cfg: RouterConfig{
				Port: 8080,
			},
			want: RouterConfig{
				Port:               8080,
				MaxRequestBodySize: defaultMaxRequestBodySize,
				ReadTimeout:        defaultHTTPServerReadTimeout,
				WriteTimeout:       defaultHTTPServerWriteTimeout,
				IdleTimeout:        defaultHTTPServerIdleTimeout,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			test.cfg.hydrateRouterDefaults()
			c.Equal(test.want, test.cfg)
		})
	}
}

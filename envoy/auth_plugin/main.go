//go:build auth_plugin

package main

import (
	"fmt"
	"os"

	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
	envoyhttp "github.com/envoyproxy/envoy/contrib/golang/filters/http/source/go/pkg/http"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"

	"github.com/buildwithgrove/auth-plugin/config"
	"github.com/buildwithgrove/auth-plugin/db"
	"github.com/buildwithgrove/auth-plugin/db/postgres"
	"github.com/buildwithgrove/auth-plugin/filter"
)

// filterName is the name of the filter that Envoy will use to identify and load the plugin
// If must match the `http_filters.typed_config.library_id` field for the Go filter in envoy.yaml
const filterName = "auth-plugin"

// CONFIG_PATH is set in the Envoy Docker image during the build process.
// It points to the mounted `.config.auth_plugin.yaml` file. See `Dockerfile.envoy`.
const envVarConfigPath = "CONFIG_PATH"

// All configuration of the plugin must be loaded in init()
// See https://github.com/envoyproxy/examples/blob/main/golang-http/simple/config.go#L16
func init() {
	logger := polyzero.NewLogger()

	configPath := os.Getenv(envVarConfigPath)
	if configPath == "" {
		panic(fmt.Sprintf("%s is not set in the environment", envVarConfigPath))
	}

	config, err := config.LoadAuthorizerPluginConfigFromYAML(configPath)
	if err != nil {
		panic(err)
	}

	dbDriver, _, err := postgres.NewPostgresDriver(config.PostgresConnectionString)
	if err != nil {
		panic(err)
	}

	cache, err := db.NewUserDataCache(dbDriver, config.CacheRefreshInterval, logger)
	if err != nil {
		panic(err)
	}

	// TODO_IMPROVE: make this configurable from the plugin config YAML
	authorizers := []filter.Authorizer{
		&filter.APIKeyAuthorizer{},
	}

	filterFactoryFunc := func(c interface{}, callbacks api.FilterCallbackHandler) api.StreamFilter {
		conf, ok := c.(*filter.EnvoyConfig)
		if !ok {
			panic("unexpected config type")
		}
		return &filter.HTTPFilter{
			Callbacks:   callbacks,
			Config:      conf,
			Authorizers: authorizers,
			Cache:       cache,
		}
	}

	envoyhttp.RegisterHttpFilterFactoryAndConfigParser(filterName, filterFactoryFunc, &filter.Parser{})
}

// func main() is only here to satisfy the Go build system
// See https://github.com/envoyproxy/examples/blob/main/golang-http/simple/config.go#L74
func main() {}

//go:build auth_plugin

package filter

import (
	"context"
	"os"

	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
	envoyhttp "github.com/envoyproxy/envoy/contrib/golang/filters/http/source/go/pkg/http"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"

	"github.com/buildwithgrove/authorizer-plugin/config"
	"github.com/buildwithgrove/authorizer-plugin/db"
	"github.com/buildwithgrove/authorizer-plugin/db/postgres"
	"github.com/buildwithgrove/authorizer-plugin/types"
)

const filterName = "authorizer-plugin"

type (
	userDataCache interface {
		GetGatewayEndpoint(ctx context.Context, userAppID types.EndpointID) (types.GatewayEndpoint, bool)
	}
)

func init() {
	logger := polyzero.NewLogger()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		panic("CONFIG_PATH is not set in the environment")
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

	filterFactoryFunc := func(c interface{}, callbacks api.FilterCallbackHandler) api.StreamFilter {
		conf, ok := c.(*envoyConfig)
		if !ok {
			panic("unexpected config type")
		}
		return &filter{
			callbacks: callbacks,
			config:    conf,
			cache:     cache,
		}
	}

	envoyhttp.RegisterHttpFilterFactoryAndConfigParser(filterName, filterFactoryFunc, &parser{})
}

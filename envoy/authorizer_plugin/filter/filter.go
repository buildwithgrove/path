//go:build authorizer_plugin

package filter

import (
	"context"
	"errors"
	"os"

	"github.com/commoddity/gonvoy"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"

	"github.com/buildwithgrove/authorizer-plugin/config"
	"github.com/buildwithgrove/authorizer-plugin/db"
	"github.com/buildwithgrove/authorizer-plugin/db/postgres"
	"github.com/buildwithgrove/authorizer-plugin/user"
)

const filterName = "authorizer-plugin"

type userDataCache interface {
	GetGatewayEndpoint(ctx context.Context, userAppID user.EndpointID) (user.GatewayEndpoint, bool)
}

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

	filterFactoryFunc := func() gonvoy.HttpFilter {
		return &Filter{cache: cache}
	}

	gonvoy.RunHttpFilter(filterName, filterFactoryFunc, gonvoy.ConfigOptions{
		FilterConfig:            new(Config),
		DisableStrictBodyAccess: true,
	})
}

/* ---------------------------------  HTTP Filter Struct -------------------------------- */

// The Filter struct handles setup of handlers for incoming service requests.
// In the case of the Authorizer Plugin, it handles adding an Authorization handler to the filter pipeline.
// This handler is responsible for authorizing incoming requests based on user data stored in an in-memory cache.
// The cache is updated with user data fetched from Postgres.
type Filter struct {
	cache userDataCache
}

var _ gonvoy.HttpFilter = &Filter{}

func (f *Filter) OnBegin(c gonvoy.RuntimeContext, ctrl gonvoy.HttpFilterController) error {

	ctrl.AddHandler(&AuthorizationHandler{
		cache: f.cache,
	})

	return nil
}

func (f *Filter) OnComplete(c gonvoy.Context) error {
	return nil
}

/* ---------------------------------  HTTP Filter Config Struct -------------------------------- */

type Config struct {
	ParentOnly     string            `json:"parent_only,omitempty"`
	RequestHeaders map[string]string `json:"request_headers,omitempty" envoy:"mergeable,preserve_root"`
	Invalid        bool              `json:"invalid,omitempty" envoy:"mergeable"`
}

func (c *Config) Validate() error {
	if c.Invalid {
		return errors.New("invalid is enabled, hence this error returned")
	}

	return nil
}

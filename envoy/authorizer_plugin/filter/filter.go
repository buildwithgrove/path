//go:build authorizer_plugin

package filter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ardikabs/gonvoy"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"

	"github.com/buildwithgrove/authorizer-plugin/db"
	"github.com/buildwithgrove/authorizer-plugin/db/postgres"
	"github.com/buildwithgrove/authorizer-plugin/user"
)

const (
	postgresConnectionString = "postgres://postgres:pgpassword@db:5432/postgres?sslmode=disable"
	cacheRefreshInterval     = 1 * time.Minute
)

type cache interface {
	GetGatewayEndpoint(ctx context.Context, userAppID user.EndpointID) (user.GatewayEndpoint, bool)
}

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

// TODO_IMPROVE: figure out how to pass to handler struct to avoid global variable
var globalCache cache

func init() {
	logger := polyzero.NewLogger()

	dbDriver, _, err := postgres.NewPostgresDriver(postgresConnectionString)
	if err != nil {
		panic(err)
	}

	cache, err := db.NewUserDataCache(dbDriver, cacheRefreshInterval, logger)
	if err != nil {
		panic(err)
	}

	globalCache = cache

	gonvoy.RunHttpFilter(new(Filter), gonvoy.ConfigOptions{
		FilterConfig:            new(Config),
		DisableStrictBodyAccess: true,
	})
}

type Filter struct{}

var _ gonvoy.HttpFilter = &Filter{}

func (f *Filter) Name() string {
	return "authorizer-plugin"
}

func (f *Filter) OnBegin(c gonvoy.RuntimeContext, ctrl gonvoy.HttpFilterController) error {
	fcfg := c.GetFilterConfig()
	cfg, ok := fcfg.(*Config)
	if !ok {
		return fmt.Errorf("unexpected configuration type %T, expecting %T", fcfg, cfg)
	}

	ctrl.AddHandler(&AuthorizationHandler{})

	return nil
}

func (f *Filter) OnComplete(c gonvoy.Context) error {
	return nil
}

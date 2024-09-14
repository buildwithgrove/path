package filter

import (
	"context"
	"fmt"
	"time"

	"github.com/ardikabs/gonvoy"
	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"

	"github.com/buildwithgrove/path-authorizer/db"
	"github.com/buildwithgrove/path-authorizer/db/postgres"
	"github.com/buildwithgrove/path-authorizer/filter/handler"
	"github.com/buildwithgrove/path-authorizer/user"
)

const (
	postgresConnectionString = "postgres://postgres:pgpassword@db:5432/postgres?sslmode=disable"
	cacheRefreshInterval     = 1 * time.Minute
)

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

type cache interface {
	GetGatewayEndpoint(ctx context.Context, userAppID user.EndpointID) (user.GatewayEndpoint, bool)
}

type Filter struct{}

var _ gonvoy.HttpFilter = &Filter{}

func (f *Filter) Name() string {
	return "path-authorizer"
}

func (f *Filter) OnBegin(c gonvoy.RuntimeContext, ctrl gonvoy.HttpFilterController) error {
	fcfg := c.GetFilterConfig()
	cfg, ok := fcfg.(*Config)
	if !ok {
		return fmt.Errorf("unexpected configuration type %T, expecting %T", fcfg, cfg)
	}

	ctrl.AddHandler(&handler.Handler{Cache: globalCache})

	return nil
}

func (f *Filter) OnComplete(c gonvoy.Context) error {
	return nil
}

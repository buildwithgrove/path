package authorizer

import (
	"context"

	"github.com/go-redis/redis_rate/v10"
	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/redis/go-redis/v9"

	reqCtx "github.com/buildwithgrove/path/request/context"
	"github.com/buildwithgrove/path/user"
)

type rateLimiter struct {
	throughputLimiter *redis_rate.Limiter
	logger            polylog.Logger
}

func newRateLimiter(redisAddr string, logger polylog.Logger) *rateLimiter {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &rateLimiter{
		throughputLimiter: redis_rate.NewLimiter(rdb),
		logger:            logger.With("component", "rate_limit_authenticator"),
	}
}

func (a *rateLimiter) authorizeRequest(ctx context.Context, reqDetails reqCtx.HTTPDetails, userApp user.GatewayEndpoint) *failedAuth {

	if throughputLimited := a.authThroughputLimit(ctx, userApp); throughputLimited != nil {
		return throughputLimited
	}

	return nil
}

func (a *rateLimiter) authThroughputLimit(ctx context.Context, userApp user.GatewayEndpoint) *failedAuth {
	throughputLimit := userApp.GetThroughputLimit()
	if throughputLimit == 0 {
		return nil
	}

	userAppThroughputLimit := redis_rate.PerSecond(throughputLimit)

	res, err := a.throughputLimiter.Allow(ctx, string(userApp.EndpointID), userAppThroughputLimit)
	if err != nil {
		a.logger.Error().Err(err).Msg("redis error: failed to check throughput limit")
		// TODO_IMPROVE - what should we do in case of redis error?
		return nil
	}

	if res.Allowed == 0 {
		return &throughputLimitExceeded
	}

	return nil
}

package authenticator

import (
	"context"

	"github.com/go-redis/redis_rate/v10"
	"github.com/pokt-network/poktroll/pkg/polylog"
	"github.com/redis/go-redis/v9"

	reqCtx "github.com/buildwithgrove/path/request/context"
	"github.com/buildwithgrove/path/user"
)

type rateLimitAuthenticator struct {
	limiter *redis_rate.Limiter
	logger  polylog.Logger
}

func newRateLimitAuthenticator(redisAddr string, logger polylog.Logger) *rateLimitAuthenticator {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &rateLimitAuthenticator{
		limiter: redis_rate.NewLimiter(rdb),
		logger:  logger.With("component", "rate_limit_authenticator"),
	}
}

func (a *rateLimitAuthenticator) authenticate(ctx context.Context, reqDetails reqCtx.HTTPDetails, userApp user.UserApp) *failedAuth {

	if throughputLimited := a.authThroughputLimit(ctx, userApp); throughputLimited != nil {
		return throughputLimited
	}

	return nil
}

func (a *rateLimitAuthenticator) authThroughputLimit(ctx context.Context, userApp user.UserApp) *failedAuth {
	if userApp.RateLimitThroughput == 0 {
		return nil
	}

	res, err := a.limiter.Allow(ctx, string(userApp.ID), redis_rate.PerSecond(userApp.RateLimitThroughput))
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

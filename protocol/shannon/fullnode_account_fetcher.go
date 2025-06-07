package shannon

import (
	"context"
	"fmt"
	"math"
	"time"

	accounttypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/pokt-network/poktroll/pkg/polylog"
	sdk "github.com/pokt-network/shannon-sdk"
	"github.com/viccon/sturdyc"
	grpcoptions "google.golang.org/grpc"
)

// ---------------- Caching Account Fetcher ----------------

// cachingPoktNodeAccountFetcher implements the PoktNodeAccountFetcher interface.
var _ sdk.PoktNodeAccountFetcher = &cachingPoktNodeAccountFetcher{}

// accountCacheTTL: No TTL for the account cache since account data never changes.
//
// time.Duration(math.MaxInt64) equals ~292 years, which is effectively infinite.
const accountCacheTTL = time.Duration(math.MaxInt64)

// accountCacheCapacity: Maximum number of entries the account cache can hold.
// This is the total capacity, not per-shard. When capacity is exceeded, the cache
// will evict a percentage of the least recently used entries from each shard.
//
// TODO_TECHDEBT(@commoddity): Revisit cache capacity based on actual # of accounts in Shannon.
const accountCacheCapacity = 200_000

// cachingPoktNodeAccountFetcher wraps an sdk.PoktNodeAccountFetcher with caching capabilities.
// It implements the same PoktNodeAccountFetcher interface but adds sturdyc caching
// in order to reduce repeated and unnecessary requests to the full node.
type cachingPoktNodeAccountFetcher struct {
	logger polylog.Logger

	// The underlying account client to delegate to when cache misses occur
	underlyingAccountClient sdk.PoktNodeAccountFetcher

	// Cache for account responses
	accountCache *sturdyc.Client[*accounttypes.QueryAccountResponse]
}

// Account implements the `sdk.PoktNodeAccountFetcher` interface with caching.
//
// See `sdk.PoktNodeAccountFetcher` interface:
//
//	https://github.com/pokt-network/shannon-sdk/blob/main/account.go#L26
//
// It matches the function signature of the CosmosSDK's account fetcher
// in order to satisfy the `sdk.PoktNodeAccountFetcher` interface.
//
// See CosmosSDK's account fetcher:
//
//	https://github.com/cosmos/cosmos-sdk/blob/main/x/auth/types/query.pb.go#L1090
func (c *cachingPoktNodeAccountFetcher) Account(
	ctx context.Context,
	req *accounttypes.QueryAccountRequest,
	opts ...grpcoptions.CallOption,
) (*accounttypes.QueryAccountResponse, error) {
	return c.accountCache.GetOrFetch(
		ctx,
		getAccountCacheKey(req.Address),
		func(fetchCtx context.Context) (*accounttypes.QueryAccountResponse, error) {
			c.logger.Debug().Str("account_key", getAccountCacheKey(req.Address)).Msgf(
				"[cachingPoktNodeAccountFetcher.Account] Making request to full node",
			)
			return c.underlyingAccountClient.Account(fetchCtx, req, opts...)
		},
	)
}

// getAccountCacheKey returns the cache key for the given account address.
// It uses the accountCacheKeyPrefix and the account address to create a unique key.
//
// eg. "account:pokt1up7zlytnmvlsuxzpzvlrta95347w322adsxslw"
func getAccountCacheKey(address string) string {
	return fmt.Sprintf("%s:%s", accountCacheKeyPrefix, address)
}

// wrapUnderlyingAccountFetcher wraps the original account fetcher with the caching
// account fetcher and replaces the lazy full node's account fetcher with the caching one.
//
// This is used to replace the lazy full node's account fetcher with the caching one.
// It is used in the NewCachingFullNode function to create a new caching full node.
func wrapUnderlyingAccountFetcher(
	logger polylog.Logger,
	lazyFullNode *lazyFullNode,
) {
	// Create the account cache, which is used to cache account responses from the full node.
	accountCache := initAccountCache()

	// Wrap the original account fetcher with the caching account fetcher
	// so that the caching account fetcher can fetch accounts from the full node.
	originalAccountFetcher := lazyFullNode.accountClient.PoktNodeAccountFetcher

	// Replace the lazy full node's account fetcher with the caching one.
	lazyFullNode.accountClient = &sdk.AccountClient{
		PoktNodeAccountFetcher: &cachingPoktNodeAccountFetcher{
			logger:                  logger,
			underlyingAccountClient: originalAccountFetcher,
			accountCache:            accountCache,
		},
	}
}

// initAccountCache initializes the account cache using SturdyC.
//
// Account data never changes, so we can cache it indefinitely.
//
// See: https://github.com/viccon/sturdyc?tab=readme-ov-file#creating-a-cache-client
func initAccountCache() *sturdyc.Client[*accounttypes.QueryAccountResponse] {
	// Create the account cache, which will be used to cache account responses.
	// This cache is effectively infinite caching for the lifetime of the application.
	// Account data never changes, so we can cache it indefinitely.
	accountCache := sturdyc.New[*accounttypes.QueryAccountResponse](
		accountCacheCapacity,
		numShards,
		accountCacheTTL,
		evictionPercentage,
	)

	return accountCache
}

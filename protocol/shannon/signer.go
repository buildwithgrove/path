package shannon

import (
	"context"
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sdk "github.com/pokt-network/shannon-sdk"
)

type signer struct {
	accountClient   sdk.AccountClient
	privateKeyHex   string
	signatureCache  *SignatureCache
}

// newSigner creates a new signer with optional signature caching
func newSigner(
	accountClient sdk.AccountClient,
	privateKeyHex string,
	logger polylog.Logger,
	cacheEnabled bool,
	cacheSize int,
) *signer {
	// Create signature cache if enabled
	var cache *SignatureCache
	if cacheEnabled {
		var err error
		cache, err = NewSignatureCache(logger, cacheSize, defaultSignatureCacheTTL, true)
		if err != nil {
			logger.Warn().Err(err).Msg("failed to create signature cache, continuing without caching")
			cache = nil
		}
	}
	
	return &signer{
		accountClient:  accountClient,
		privateKeyHex:  privateKeyHex,
		signatureCache: cache,
	}
}

func (s *signer) SignRelayRequest(req *servicetypes.RelayRequest, app apptypes.Application) (*servicetypes.RelayRequest, error) {
	// If cache is available, use it
	if s.signatureCache != nil {
		return s.signatureCache.GetOrCompute(req, app, func() (*servicetypes.RelayRequest, error) {
			return s.computeSignature(req, app)
		})
	}
	
	// No cache, compute directly
	return s.computeSignature(req, app)
}

// computeSignature performs the actual signature computation (expensive operation)
func (s *signer) computeSignature(req *servicetypes.RelayRequest, app apptypes.Application) (*servicetypes.RelayRequest, error) {
	ring := sdk.ApplicationRing{
		Application:      app,
		PublicKeyFetcher: &s.accountClient,
	}

	sdkSigner := sdk.Signer{PrivateKeyHex: s.privateKeyHex}
	req, err := sdkSigner.Sign(context.Background(), req, ring)
	if err != nil {
		return nil, fmt.Errorf("SignRequest: error signing relay request: %w", err)
	}

	return req, nil
}

// GetCacheStats returns the signature cache statistics
func (s *signer) GetCacheStats() *SignatureCacheStats {
	if s.signatureCache == nil {
		return nil
	}
	stats := s.signatureCache.GetStats()
	return &stats
}

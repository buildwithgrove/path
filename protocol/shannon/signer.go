package shannon

import (
	"context"
	"fmt"

	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sdk "github.com/pokt-network/shannon-sdk"
)

type signer struct {
	accountClient sdk.AccountClient
	privateKeyHex string
}

func (s *signer) SignRelayRequest(req *servicetypes.RelayRequest, app apptypes.Application) (*servicetypes.RelayRequest, error) {
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

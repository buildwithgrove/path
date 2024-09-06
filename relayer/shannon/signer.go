package shannon

import (
	"context"
	"fmt"

	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sdk "github.com/pokt-network/shannon-sdk"
)

func newSigner(privateKeyHex string, config GRPCConfig) (*signer, error) {
	conn, err := connectGRPC(config)
	if err != nil {
		return nil, fmt.Errorf("newSigner: could not create new Shannon Signer. Error establishing grpc connection to url %s: %w", config.HostPort, err)
	}

	return &signer{
		privateKeyHex: privateKeyHex,
		accountClient: sdk.AccountClient{PoktNodeAccountFetcher: sdk.NewPoktNodeAccountFetcher(conn)},
	}, nil
}

type signer struct {
	accountClient sdk.AccountClient
	privateKeyHex string
}

func (s *signer) SignRequest(req *servicetypes.RelayRequest, app apptypes.Application) (*servicetypes.RelayRequest, error) {
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

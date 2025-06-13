package protocol

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"

	sdk "github.com/pokt-network/shannon-sdk"
)

// TODO_MVP(@adshmh): add a README based on the following notion doc:
// https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5
//
// GatewayMode represents the operation mode of the gateway that is using a relaying protcol for serving user requests.
// It is defined in the `protocol` package for:
// a. Consistency: protocol package defines all key concepts related to a relaying protocol.
// b. Clarity: make it clear that is a protocol-level concern to be addressed by all protocol implementations.
//
// As of now, the GatewayMode can be one of the following (see the comment attached to each mode for details).
// 1. Centralized
// 2. Delegated
// 3. Permissionless
type GatewayMode string

const (
	GatewayModeCentralized    = "centralized"
	GatewayModeDelegated      = "delegated"
	GatewayModePermissionless = "permissionless"
)

// TODO - move this to its own file
type FullNode interface {
	// GetApp returns the onchain application matching the application address
	GetApp(ctx context.Context, appAddr string) (*apptypes.Application, error)

	// GetSession returns the latest session matching the supplied service+app combination.
	// Sessions are solely used for sending relays, and therefore only the latest session for any service+app combination is needed.
	// Note: Shannon returns the latest session for a service+app combination if no blockHeight is provided.
	GetSession(ctx context.Context, serviceID sdk.ServiceID, appAddr string) (sessiontypes.Session, error)

	// GetAccountPubKey returns the account public key for the given address.
	// The cache has no TTL, so the public key is cached indefinitely.
	GetAccountPubKey(ctx context.Context, address string) (cryptotypes.PubKey, error)

	// ValidateRelayResponse validates the raw bytes returned from an endpoint (in response to a relay request) and returns the parsed response.
	ValidateRelayResponse(ctx context.Context, supplierAddr sdk.SupplierAddress, responseBz []byte) (*servicetypes.RelayResponse, error)

	// IsHealthy returns true if the FullNode instance is healthy.
	// A LazyFullNode will always return true.
	// A CachingFullNode will return true if it has data in app and session caches.
	IsHealthy() bool
}

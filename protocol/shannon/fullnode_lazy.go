package shannon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
	sdktypes "github.com/pokt-network/shannon-sdk/types"

	"github.com/buildwithgrove/path/protocol"
)

// The Shannon FullNode interface is implemented by the lazyFullNode struct below.
//
// A lazyFullNode queries the onchain data for every data item it needs to do an action (e.g. serve a relay request, etc).
// This is done to enable supporting short block times (a few seconds), by avoiding caching
// which can result in failures due to stale data in the cache.
var _ FullNode = &lazyFullNode{}

// TODO_MVP(@adshmh): Rename `lazyFullNode`: this struct does not perform any caching and should be named accordingly.
//
// LazyFullNode: default implementation of a full node for the Shannon.
//
// Key differences from a caching full node:
// - Intentionally avoids caching:
//   - Enables support for short block times (e.g. LocalNet)
//   - Use CachingFullNode struct if caching is desired for performance
type lazyFullNode struct {
	// logger polylog.Logger

	appClient     *sdk.ApplicationClient
	sessionClient *sdk.SessionClient
	blockClient   *sdk.BlockClient
	accountClient *sdk.AccountClient
}

// NewLazyFullNode builds and returns a LazyFullNode using the provided configuration.
func NewLazyFullNode(config FullNodeConfig) (*lazyFullNode, error) {
	blockClient, err := newBlockClient(config.RpcURL)
	if err != nil {
		return nil, fmt.Errorf("NewSdk: error creating new Shannon block client at URL %s: %w", config.RpcURL, err)
	}

	config.GRPCConfig = config.GRPCConfig.hydrateDefaults()

	sessionClient, err := newSessionClient(config.GRPCConfig)
	if err != nil {
		return nil, fmt.Errorf("NewSdk: error creating new Shannon session client using URL %s: %w", config.GRPCConfig.HostPort, err)
	}

	appClient, err := newAppClient(config.GRPCConfig)
	if err != nil {
		return nil, fmt.Errorf("NewSdk: error creating new GRPC connection at url %s: %w", config.GRPCConfig.HostPort, err)
	}

	accountClient, err := newAccClient(config.GRPCConfig)
	if err != nil {
		return nil, fmt.Errorf("NewSdk: error creating new account client using url %s: %w", config.GRPCConfig.HostPort, err)
	}

	fullNode := &lazyFullNode{
		sessionClient: sessionClient,
		appClient:     appClient,
		blockClient:   blockClient,
		accountClient: accountClient,
	}

	return fullNode, nil
}

// GetApp:
// - Returns the onchain application matching the supplied application address.
// - Required to fulfill the FullNode interface.
func (lfn *lazyFullNode) GetApp(ctx context.Context, appAddr string) (*apptypes.Application, error) {
	app, err := lfn.appClient.GetApplication(ctx, appAddr)
	return &app, err
}

// GetSession:
// - Uses the Shannon SDK to fetch a session for the (serviceID, appAddr) combination.
// - Required to fulfill the FullNode interface.
func (lfn *lazyFullNode) GetSession(
	ctx context.Context,
	serviceID protocol.ServiceID,
	appAddr string,
) (sessiontypes.Session, error) {
	session, err := lfn.sessionClient.GetSession(
		ctx,
		appAddr,
		string(serviceID),
		0,
	)

	if err != nil {
		return sessiontypes.Session{},
			fmt.Errorf("GetSession: error getting the session for service %s app %s: %w",
				serviceID, appAddr, err,
			)
	}

	if session == nil {
		return sessiontypes.Session{},
			fmt.Errorf("GetSession: got nil session for service %s app %s: %w",
				serviceID, appAddr, err,
			)
	}

	return *session, nil
}

// ValidateRelayResponse:
// - Validates the raw response bytes received from an endpoint.
// - Uses the SDK and the account client for validation.
func (lfn *lazyFullNode) ValidateRelayResponse(supplierAddr sdk.SupplierAddress, responseBz []byte) (*servicetypes.RelayResponse, error) {
	return sdk.ValidateRelayResponse(
		context.Background(),
		supplierAddr,
		responseBz,
		lfn.accountClient,
	)
}

// IsHealthy:
// - Always returns true for a LazyFullNode.
// - Required to fulfill the FullNode interface.
func (lfn *lazyFullNode) IsHealthy() bool {
	return true
}

// GetAccountClient:
// - Returns the account client created by the lazy fullnode.
// - Used to create relay request signers.
func (lfn *lazyFullNode) GetAccountClient() *sdk.AccountClient {
	return lfn.accountClient
}

// serviceRequestPayload:
// - Contents of the request received by the underlying service's API server.

func shannonJsonRpcHttpRequest(serviceRequestPayload []byte, url string) (*http.Request, error) {
	jsonRpcServiceReq, err := http.NewRequest(http.MethodPost, url, io.NopCloser(bytes.NewReader(serviceRequestPayload)))
	if err != nil {
		return nil, fmt.Errorf("shannonJsonRpcHttpRequest: failed to create a new HTTP request for url %s: %w", url, err)
	}

	jsonRpcServiceReq.Header.Set("Content-Type", "application/json")
	return jsonRpcServiceReq, nil
}

func embedHttpRequest(reqToEmbed *http.Request) (*servicetypes.RelayRequest, error) {
	_, reqToEmbedBz, err := sdktypes.SerializeHTTPRequest(reqToEmbed)
	if err != nil {
		return nil, fmt.Errorf("embedHttpRequest: failed to Serialize HTTP Request for URL %s: %w", reqToEmbed.URL, err)
	}

	return &servicetypes.RelayRequest{
		Payload: reqToEmbedBz,
	}, nil
}

// TODO_IMPROVE: consider enhancing the service or RelayRequest/RelayResponse types in poktroll repo (see links below) to perform
// Serialization/Deserialization of the payload.
//
// Benefits:
// - Makes code easier to read and less error-prone by consolidating serialization/deserialization in a single source (e.g. relay.go).
//
// Current behavior:
// - Relay miner serializes the HTTP response received from the proxied service (see link below).
// - Deserialization is handled here (see sdktypes.DeserializeHTTPResponse below).
//
// Links:
// - Relay miner serializing the service response:
//   https://github.com/pokt-network/poktroll/blob/e5024ea5d28cc94d09e531f84701a85cefb9d56f/pkg/relayer/proxy/synchronous.go#L361-L363
// - Relay response validation (potential package for serialization/deserialization):
//   https://github.com/pokt-network/poktroll/blob/e5024ea5d28cc94d09e531f84701a85cefb9d56f/x/service/types/relay.go#L68
//
// deserializeRelayResponse:
// - Uses the Shannon sdk to deserialize the relay response payload received from an endpoint into a protocol.Response.
// - Required because the relay miner (the endpoint serving the relay) returns the HTTP response in serialized format in its payload.

func deserializeRelayResponse(bz []byte) (protocol.Response, error) {
	poktHttpResponse, err := sdktypes.DeserializeHTTPResponse(bz)
	if err != nil {
		return protocol.Response{}, err
	}

	return protocol.Response{
		Bytes:          poktHttpResponse.BodyBz,
		HTTPStatusCode: int(poktHttpResponse.StatusCode),
	}, nil
}

func newSessionClient(config GRPCConfig) (*sdk.SessionClient, error) {
	conn, err := connectGRPC(config)
	if err != nil {
		return nil, fmt.Errorf("could not create new Shannon session client: error establishing grpc connection to %s: %w", config.HostPort, err)
	}

	return &sdk.SessionClient{PoktNodeSessionFetcher: sdk.NewPoktNodeSessionFetcher(conn)}, nil
}

func newBlockClient(fullNodeURL string) (*sdk.BlockClient, error) {
	_, err := url.Parse(fullNodeURL)
	if err != nil {
		return nil, fmt.Errorf("newBlockClient: error parsing url %s: %w", fullNodeURL, err)
	}

	nodeStatusFetcher, err := sdk.NewPoktNodeStatusFetcher(fullNodeURL)
	if err != nil {
		return nil, fmt.Errorf("newBlockClient: error connecting to a full node %s: %w", fullNodeURL, err)
	}

	return &sdk.BlockClient{PoktNodeStatusFetcher: nodeStatusFetcher}, nil
}

func newAppClient(config GRPCConfig) (*sdk.ApplicationClient, error) {
	appConn, err := connectGRPC(config)
	if err != nil {
		return nil, fmt.Errorf("NewSdk: error creating new GRPC connection at url %s: %w", config.HostPort, err)
	}

	return &sdk.ApplicationClient{QueryClient: apptypes.NewQueryClient(appConn)}, nil
}

func newAccClient(config GRPCConfig) (*sdk.AccountClient, error) {
	conn, err := connectGRPC(config)
	if err != nil {
		return nil, fmt.Errorf("newAccClient: error creating new GRPC connection for account client at url %s: %w", config.HostPort, err)
	}

	return &sdk.AccountClient{PoktNodeAccountFetcher: sdk.NewPoktNodeAccountFetcher(conn)}, nil
}

package shannon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
	sdktypes "github.com/pokt-network/shannon-sdk/types"

	"github.com/buildwithgrove/path/protocol"
)

// The Shannon Relayer's FullNode interface is implemented by the LazyFullNode struct below,
// which provides the full node capabilities required by the Shannon relayer.
// A LazyFullNode queries the onchain data for every data item it needs to serve a relay request, e.g. applications staked for a service.
// This is done to enable supporting short block times (a few seconds), by avoiding caching which can result in failures due to stale
// data in the cache.
//
// A properly initialized fullNode struct can:
// 1. Return the onchain apps matching a service ID.
// 2. Fetch a session for a (service,app) combination.
// 3. Send a relay, corresponding to a specific session, to an endpoint.
var _ FullNode = &LazyFullNode{}

// NewLazyFullNode builds and returns a LazyFullNode using the provided configuration.
func NewLazyFullNode(config FullNodeConfig, logger polylog.Logger) (*LazyFullNode, error) {
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

	lazyFullNode := &LazyFullNode{
		sessionClient: sessionClient,
		appClient:     appClient,
		blockClient:   blockClient,
		accountClient: accountClient,

		logger: logger,
	}

	return lazyFullNode, nil
}

// LazyFullNode provides the default implementation of a full node required by the Shannon relayer.
// The key differences between a lazy and full node are:
// 1. Lazy node intentionally avoids caching.
//   - This allows supporting short block times (e.g. LocalNet)
//   - CachingFullNode struct can be used instead if caching is desired for performance reasons
type LazyFullNode struct {
	appClient     *sdk.ApplicationClient
	sessionClient *sdk.SessionClient
	blockClient   *sdk.BlockClient
	accountClient *sdk.AccountClient

	permittedAppFilter permittedAppFilter

	logger polylog.Logger
}

// SetPermittedAppFilter sets the permitted app filter for the protocol instance.
func (lfn *LazyFullNode) SetPermittedAppFilter(permittedAppFilter permittedAppFilter) {
	lfn.permittedAppFilter = permittedAppFilter
}

// GetServiceEndpoints returns the set of endpoints matching the supplied service ID.
// It is required to fulfill the FullNode interface.
func (lfn *LazyFullNode) GetServiceEndpoints(serviceID protocol.ServiceID, req *http.Request) (map[protocol.EndpointAddr]endpoint, error) {
	allApps, err := lfn.getAllAppsForRequest(context.TODO(), req)
	if err != nil {
		return nil, err
	}

	// use a filter to drop any apps that are not staked for the service matching the supplied service ID.
	appsServiceMap, err := lfn.buildAppsServiceMap(allApps, serviceAppFilter(serviceID))
	if err != nil {
		return nil, err
	}

	// convert the map of service ID to application which is returned from the previous method call, into a slice for easier processing.
	apps, err := lfn.getAppsUniqueEndpoints(serviceID, appsServiceMap[serviceID])
	if err != nil {
		return nil, err
	}

	return apps, nil
}

// GetSession uses the Shannon SDK to fetch a session for the (serviceID, appAddr) combination.
// It is required to fulfill the FullNode interface.
func (lfn *LazyFullNode) GetSession(serviceID protocol.ServiceID, appAddr string) (sessiontypes.Session, error) {
	session, err := lfn.sessionClient.GetSession(
		context.Background(),
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

// ValidateRelayResponse validates the raw response bytes received from an endpoint using the SDK and the account client.
func (lfn *LazyFullNode) ValidateRelayResponse(supplierAddr sdk.SupplierAddress, responseBz []byte) (*servicetypes.RelayResponse, error) {
	return sdk.ValidateRelayResponse(
		context.Background(),
		supplierAddr,
		responseBz,
		lfn.accountClient,
	)
}

// IsHealthy always returns true for a LazyFullNode.
// It is required to fulfill the FullNode interface.
func (lfn *LazyFullNode) IsHealthy() bool {
	return true
}

// GetAccountClient returns the account client created by the lazy fullnode.
// It is used to create relay request signers.
func (lfn *LazyFullNode) GetAccountClient() *sdk.AccountClient {
	return lfn.accountClient
}

// TODO_FUTURE(@adshmh): Find a more optimized way of handling an overlap among endpoints
// matching multiple sessions of apps delegating to the gateway.
//
// getAppsUniqueEndpoints returns a map of all endpoints which match the provided service ID and pass the supplied app filter.
// If an endpoint matches a service ID through multiple apps/sessions, only a single entry
// matching one of the apps/sessions is returned.
func (lfn *LazyFullNode) getAppsUniqueEndpoints(serviceID protocol.ServiceID, apps []apptypes.Application) (map[protocol.EndpointAddr]endpoint, error) {
	endpoints := make(map[protocol.EndpointAddr]endpoint)

	for _, app := range apps {
		session, err := lfn.GetSession(serviceID, app.Address)
		if err != nil {
			return nil, fmt.Errorf("getAppsUniqueEndpoints: could not get the session for service %s app %s", serviceID, app.Address)
		}

		appEndpoints, err := endpointsFromSession(session)
		if err != nil {
			return nil, fmt.Errorf("getAppsUniqueEndpoints: error getting all endpoints for app %s session %s: %w", app.Address, session.SessionId, err)
		}

		for endpointAddr, endpoint := range appEndpoints {
			endpoints[endpointAddr] = endpoint
		}
	}

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("getAppsUniqueEndpoints: no endpoints found for service %s", serviceID)
	}

	return endpoints, nil
}

// buildAppsServiceIdx builds a map of serviceIDs to the corresponding onchain apps.
func (lfn *LazyFullNode) buildAppsServiceMap(onchainApps []apptypes.Application, filterFn appFilterFn) (map[protocol.ServiceID][]apptypes.Application, error) {
	appData := make(map[protocol.ServiceID][]apptypes.Application)

	for _, onchainApp := range onchainApps {
		logger := lfn.logger.With("address", onchainApp.Address)

		if len(onchainApp.ServiceConfigs) == 0 {
			logger.Warn().Msg("buildAppsServiceMap: app has no services specified onchain. Skipping the app.")
			continue
		}

		for _, svcCfg := range onchainApp.ServiceConfigs {
			if svcCfg.ServiceId == "" {
				logger.Warn().Msg("buildAppsServiceMap: app has empty serviceId item in service config.")
				continue
			}

			if filterFn != nil && !filterFn(onchainApp, protocol.ServiceID(svcCfg.ServiceId)) {
				continue
			}

			serviceID := protocol.ServiceID(svcCfg.ServiceId)
			appData[serviceID] = append(appData[serviceID], onchainApp)
		}
	}

	if len(appData) == 0 {
		return nil, fmt.Errorf("buildAppsServiceMap: no apps found")
	}

	return appData, nil
}

// getAllAppsForRequest returns all the onchain apps; it is used by the lazy full node to fetch apps for a request.
// TODO_MVP(@adshmh): query the onchain data for the gateway address to confirm it is valid and return an error if not.
func (lfn *LazyFullNode) getAllAppsForRequest(ctx context.Context, req *http.Request) ([]apptypes.Application, error) {
	// TODO_MVP(@adshmh): remove this once poktroll supports querying the onchain apps.
	// More specifically, support for the following criteria is required as of now:
	// 1. Apps matching a specific service ID
	// 2. Apps delegating to a gateway address.
	appsData, err := lfn.appClient.GetAllApplications(ctx)
	if err != nil {
		return nil, fmt.Errorf("getAllApps: error getting all applications: %w", err)
	}

	// The request is passed to filterPermittedApps to filter apps based on the gateway mode.
	// - In `centralized` mode the apps are filtering based on the gateway's owned apps.
	// - In `delegated` mode the apps are filtering based on the app address specified in the HTTP request's headers.
	return lfn.filterPermittedApps(appsData, req), nil
}

// filterPermittedApps filters the apps based on the permittedAppFilter, which is determined by the gateway mode.
// TODO_MVP(@adshmh): once poktroll support querying the onchain apps, this function should be removed.
func (lfn *LazyFullNode) filterPermittedApps(apps []apptypes.Application, req *http.Request) []apptypes.Application {
	var filteredApps []apptypes.Application

	for _, app := range apps {
		if errSelectingApp := lfn.permittedAppFilter(&app, req); errSelectingApp != nil {
			lfn.logger.Info().Err(errSelectingApp).Str("app_address", app.Address).
				Msg("fetchApps: app filter rejected the app: skipping the app")
			continue
		}

		filteredApps = append(filteredApps, app)
	}

	return filteredApps
}

// serviceRequestPayload is the contents of the request received by the underlying service's API server.
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

// TODO_IMPROVE: consider enhancing the service or RelayRequest/RelayResponse types in poktroll repo (link below) to perform
// Serialization/Deserialization of the payload. This will make the code easier to read and less error prone as a single
// source, e.g. the relay.go file linked below, would be responsible for both operations.
// Currently, the relay miner serializes the HTTP response received from the service it proxies (link below), while the
// deserialization needs to take place here (see the call to sdktypes.DeserializeHTTPResponse below).

// Link to relay miner serializing the service response:
// https://github.com/pokt-network/poktroll/blob/e5024ea5d28cc94d09e531f84701a85cefb9d56f/pkg/relayer/proxy/synchronous.go#L361-L363
//
// Link to relay response validation, as a potentially good package fit for performing serialization/deserialization of relay request/response.
// https://github.com/pokt-network/poktroll/blob/e5024ea5d28cc94d09e531f84701a85cefb9d56f/x/service/types/relay.go#L68
//
// deserializeRelayResponse uses the Shannon sdk to deserialize the relay response payload
// received from an endpoint into a protocol.Response. This is necessary since the relay miner, i.e. the endpoint
// that serves the relay, returns the HTTP response in serialized format in its payload.
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

// appFilterFn represents a filter that determines whether an app should be included.
// it is mainly used to return the apps matching a specific service ID.
type appFilterFn func(apptypes.Application, protocol.ServiceID) bool

// serviceAppFilter is an app filtering function that drops any applications which does not match the supplied service ID.
func serviceAppFilter(selectedServiceID protocol.ServiceID) appFilterFn {
	return func(_ apptypes.Application, serviceID protocol.ServiceID) bool {
		return serviceID == selectedServiceID
	}
}

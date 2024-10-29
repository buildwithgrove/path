package shannon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
	sdktypes "github.com/pokt-network/shannon-sdk/types"

	"github.com/buildwithgrove/path/relayer"
)

// The Shannon Relayer's FullNode interface is implemented by the LazyFullNode struct below,
// which provides the full node capabilities required by the Shannon relayer.
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

	signer, err := newSigner(config.GatewayPrivateKey, config.GRPCConfig)
	if err != nil {
		return nil, fmt.Errorf("NewSdk: error creating new signer at url %s: %w", config.GRPCConfig.HostPort, err)
	}

	lazyFullNode := &LazyFullNode{
		gatewayAddress: config.GatewayAddress,
		delegatedApps:  config.DelegatedApps,

		sessionClient: sessionClient,
		appClient:     appClient,
		blockClient:   blockClient,
		accountClient: accountClient,
		signer:        signer,

		logger: logger,
	}

	return lazyFullNode, nil
}

// LazyFullNode provides the default implementation of a full node required by the Shannon relayer.
// The key differences between a lazy and full node are:
// 1. Lazy node intentionally avoids caching.
// 	- This allows supporting short block times (e.g. LocalNet)
//      - CachingFullNode struct can be used instead if caching is desired for performance reasons
type LazyFullNode struct {
	// gatewayAddress is used by the SDK for selecting onchain applications which have delegated to the gateway.
	// The gateway can only sign relays on behalf of an application if the application has an active delegation to it.
	gatewayAddress string
	// TODO_UPNEXT(@adshmh): use private keys of owned apps.
	delegatedApps []string

	appClient     *sdk.ApplicationClient
	sessionClient *sdk.SessionClient
	blockClient   *sdk.BlockClient
	accountClient *sdk.AccountClient
	signer        *signer

	logger polylog.Logger
}

func (lfn *LazyFullNode) GetServiceApps(serviceID relayer.ServiceID) ([]apptypes.Application, error) {
	allApps, err := lfn.getAllApps(context.TODO())
	if err != nil {
		return nil, err
	}

	appsServiceMap, err := lfn.buildAppsServiceMap(allApps, serviceAppFilter(serviceID))
	if err != nil {
		return nil, err
	}

	var apps []apptypes.Application
	for appsSvcID, svcApps := range appsServiceMap {
		if appsSvcID != serviceID {
			continue
		}

		apps = append(apps, svcApps...)
	}

	return apps, nil
}

func (lfn *LazyFullNode) GetAllServicesApps() (map[relayer.ServiceID][]apptypes.Application, error) {
	allApps, err := lfn.getAllApps(context.TODO())
	if err != nil {
		return nil, err
	}
	return lfn.buildAppsServiceMap(allApps, nil)
}

func (lfn *LazyFullNode) GetSession(serviceID relayer.ServiceID, appAddr string) (sessiontypes.Session, error) {
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

// TODO_IMPROVE: split this function into build/sign/send/verify stages.
func (lfn *LazyFullNode) SendRelay(app apptypes.Application, session sessiontypes.Session, endpoint endpoint, payload relayer.Payload) (*servicetypes.RelayResponse, error) {
	// TODO_TECHDEBT: need to select the correct underlying request (HTTP, etc.) based on the selected service.
	jsonRpcHttpReq, err := shannonJsonRpcHttpRequest([]byte(payload.Data), endpoint.url)
	if err != nil {
		return nil, fmt.Errorf("error building a JSONRPC HTTP request for url %s: %w", endpoint.url, err)
	}

	relayRequest, err := embedHttpRequest(jsonRpcHttpReq)
	if err != nil {
		return nil, fmt.Errorf("error embedding a JSONRPC HTTP request for url %s: %w", endpoint.url, err)
	}

	// TODO_TECHDEBT: use the new `FilteredSession` struct provided by the Shannon SDK to get the session and the endpoint.
	relayRequest.Meta = servicetypes.RelayRequestMetadata{
		SessionHeader:           session.Header,
		SupplierOperatorAddress: string(endpoint.supplier),
	}

	req, err := lfn.signer.SignRequest(relayRequest, app)
	if err != nil {
		return nil, fmt.Errorf("relay: error signing the relay request for app %s: %w", app.Address, err)
	}

	ctxWithTimeout, cancelFn := context.WithTimeout(context.Background(), time.Duration(payload.TimeoutMillisec)*time.Millisecond)
	defer cancelFn()

	responseBz, err := sendHttpRelay(ctxWithTimeout, endpoint.url, req)
	if err != nil {
		return nil, fmt.Errorf("relay: error sending request to endpoint %s: %w", endpoint.url, err)
	}

	// Validate the response
	response, err := sdk.ValidateRelayResponse(
		context.Background(),
		sdk.SupplierAddress(endpoint.supplier),
		responseBz,
		lfn.accountClient,
	)
	if err != nil {
		return nil, fmt.Errorf("relay: error verifying the relay response for app %s, endpoint %s: %w", app.Address, endpoint.url, err)
	}

	return response, nil
}

// IsHealthy always returns true for a LazyFullNode.
func (lfn *LazyFullNode) IsHealthy() bool {
	return true
}

// buildAppsServiceIdx builds a map of serviceIDs to the corresponding onchain apps.
func (lfn *LazyFullNode) buildAppsServiceMap(onchainApps []apptypes.Application, filterFn appFilterFn) (map[relayer.ServiceID][]apptypes.Application, error) {
	appData := make(map[relayer.ServiceID][]apptypes.Application)
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

			if filterFn != nil && !filterFn(onchainApp, relayer.ServiceID(svcCfg.ServiceId)) {
				continue
			}

			serviceID := relayer.ServiceID(svcCfg.ServiceId)
			appData[serviceID] = append(appData[serviceID], onchainApp)
		}
	}

	if len(appData) == 0 {
		return nil, fmt.Errorf("buildAppsServiceMap: no apps found.")
	}

	return appData, nil
}

// TODO_UPNEXT(@adshmh): cross-reference onchain apps against the configured apps' private keys.
// An onchain app should be used for sending relays if and only if it meets the following criteria:
// 1. It has delegated to the gateway.
// 2. Its private key is present in the configuration.
//
// getAllApps returns the onchain apps that have active delegations to the gateway.
func (lfn *LazyFullNode) getAllApps(ctx context.Context) ([]apptypes.Application, error) {
	// TODO_TECHDEBT: query the onchain data for the gateway address to confirm it is valid and return an error if not.

	var apps []apptypes.Application
	for _, appAddr := range lfn.delegatedApps {
		onchainApp, err := lfn.appClient.GetApplication(ctx, appAddr)
		if err != nil {
			lfn.logger.Error().Msgf("GetApps: SDK returned error when getting application %s: %v", appAddr, err)
			continue
		}

		if !slices.Contains(onchainApp.DelegateeGatewayAddresses, lfn.gatewayAddress) {
			lfn.logger.Warn().Msgf("GetApps: Application %s is not delegated to Gateway", onchainApp.Address)
			continue
		}

		apps = append(apps, onchainApp)
	}

	return apps, nil
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
// received from an endpoint into a relayer.Response. This is necessary since the relay miner, i.e. the endpoint
// that serves the relay, returns the HTTP response in serialized format in its payload.
func deserializeRelayResponse(bz []byte) (relayer.Response, error) {
	poktHttpResponse, err := sdktypes.DeserializeHTTPResponse(bz)
	if err != nil {
		return relayer.Response{}, err
	}

	return relayer.Response{
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
type appFilterFn func(apptypes.Application, relayer.ServiceID) bool

func serviceAppFilter(selectedServiceID relayer.ServiceID) appFilterFn {
	return func(_ apptypes.Application, serviceID relayer.ServiceID) bool {
		return serviceID == selectedServiceID
	}
}

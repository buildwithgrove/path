package shannon

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
	sdktypes "github.com/pokt-network/shannon-sdk/types"

	"github.com/buildwithgrove/path/relayer"
)

const (
	gatewayPrivateKeyLength = 64
	addressLength           = 43
)

// The Shannon Relayer's FullNode interface is implemented by the fullNode struct below,
// which provides the full node capabilities required by the Shannon relayer.
// A properly initialized fullNode struct can return the latest block height,
// fetch a session for a service+app combination, and send a relay to an endpoint.
var _ FullNode = &fullNode{}

var (
	ErrShannonInvalidGatewayPrivateKey = errors.New("invalid shannon gateway private key")
	ErrShannonInvalidGatewayAddress    = errors.New("invalid shannon gateway address")
	ErrShannonInvalidNodeUrl           = errors.New("invalid shannon node URL")
	ErrShannonInvalidGrpcHostPort      = errors.New("invalid shannon grpc host:port")
)

func NewFullNode(config FullNodeConfig, logger polylog.Logger) (FullNode, error) {
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

	fullNode := &fullNode{
		gatewayAddress: config.GatewayAddress,
		delegatedApps:  config.DelegatedApps,

		sessionClient: sessionClient,
		appClient:     appClient,
		blockClient:   blockClient,
		accountClient: accountClient,
		signer:        signer,

		logger: logger,
	}

	return fullNode, nil
}

type (
	// TODO_DISCUSS: move this (and the morse FullNodeConfig) to the config package?
	FullNodeConfig struct {
		RpcURL            string     `yaml:"rpc_url"`
		GRPCConfig        GRPCConfig `yaml:"grpc_config"`
		GatewayAddress    string     `yaml:"gateway_address"`
		GatewayPrivateKey string     `yaml:"gateway_private_key"`
		// A list of addresses of onchain Applications delegated to the Gateway.
		DelegatedApps []string `yaml:"delegated_app_addresses"`
	}

	GRPCConfig struct {
		HostPort          string        `yaml:"host_port"`
		Insecure          bool          `yaml:"insecure"`
		BackoffBaseDelay  time.Duration `yaml:"backoff_base_delay"`
		BackoffMaxDelay   time.Duration `yaml:"backoff_max_delay"`
		MinConnectTimeout time.Duration `yaml:"min_connect_timeout"`
		KeepAliveTime     time.Duration `yaml:"keep_alive_time"`
		KeepAliveTimeout  time.Duration `yaml:"keep_alive_timeout"`
	}
)

// TODO_IMPROVE: move this to the config package?
func (c FullNodeConfig) Validate() error {
	if len(c.GatewayPrivateKey) != gatewayPrivateKeyLength {
		return ErrShannonInvalidGatewayPrivateKey
	}
	if len(c.GatewayAddress) != addressLength {
		return ErrShannonInvalidGatewayAddress
	}
	if !strings.HasPrefix(c.GatewayAddress, "pokt1") {
		return ErrShannonInvalidGatewayAddress
	}
	if !isValidUrl(c.RpcURL, false) {
		return ErrShannonInvalidNodeUrl
	}
	if !isValidGrpcHostPort(c.GRPCConfig.HostPort) {
		return ErrShannonInvalidGrpcHostPort
	}
	for _, addr := range c.DelegatedApps {
		if len(addr) != addressLength {
			return fmt.Errorf("invalid delegated app address: %s", addr)
		}
	}
	return nil
}

// TODO_IMPROVE: move this to the config package?
const (
	defaultBackoffBaseDelay  = 1 * time.Second
	defaultBackoffMaxDelay   = 120 * time.Second
	defaultMinConnectTimeout = 20 * time.Second
	defaultKeepAliveTime     = 20 * time.Second
	defaultKeepAliveTimeout  = 20 * time.Second
)

// TODO_IMPROVE: move this to the config package?
func (c *GRPCConfig) hydrateDefaults() GRPCConfig {
	if c.BackoffBaseDelay == 0 {
		c.BackoffBaseDelay = defaultBackoffBaseDelay
	}
	if c.BackoffMaxDelay == 0 {
		c.BackoffMaxDelay = defaultBackoffMaxDelay
	}
	if c.MinConnectTimeout == 0 {
		c.MinConnectTimeout = defaultMinConnectTimeout
	}
	if c.KeepAliveTime == 0 {
		c.KeepAliveTime = defaultKeepAliveTime
	}
	if c.KeepAliveTimeout == 0 {
		c.KeepAliveTimeout = defaultKeepAliveTimeout
	}
	return *c
}

// fullNode provides the default implementation of a full node required by the Shannon relayer.
type fullNode struct {
	// gatewayAddress is used by the SDK for selecting onchain applications which have delegated to the gateway.
	// The gateway can only sign relays on behalf of an application if the application has an active delegation to it.
	gatewayAddress string
	// TODO_COMMENT
	delegatedApps []string

	appClient     *sdk.ApplicationClient
	sessionClient *sdk.SessionClient
	blockClient   *sdk.BlockClient
	accountClient *sdk.AccountClient
	signer        *signer

	logger polylog.Logger
}

func (s *fullNode) LatestBlockHeight() (int64, error) {
	return s.blockClient.LatestBlockHeight(context.Background())
}

func (s *fullNode) GetSession(serviceID, appAddr string, blockHeight int64) (sessiontypes.Session, error) {
	session, err := s.sessionClient.GetSession(
		context.Background(),
		appAddr,
		serviceID,
		blockHeight,
	)

	if err != nil {
		return sessiontypes.Session{},
			fmt.Errorf("GetSession: error getting the session for service %s app %s blockheight %d: %w",
				serviceID, appAddr, blockHeight, err,
			)
	}

	if session == nil {
		return sessiontypes.Session{},
			fmt.Errorf("GetSession: got nil session for service %s app %s blockheight %d: %w",
				serviceID, appAddr, blockHeight, err,
			)
	}

	return *session, nil
}

// GetApps returns the onchain apps that have active delegations to the gateway.
func (s *fullNode) GetApps(ctx context.Context) ([]apptypes.Application, error) {
	// TODO_TECHDEBT: query the onchain data for the gateway address to confirm it is valid and return an error if not.

	var apps []apptypes.Application
	for _, appAddr := range s.delegatedApps {
		onchainApp, err := s.appClient.GetApplication(ctx, appAddr)
		if err != nil {
			s.logger.Error().Msgf("GetApps: SDK returned error when getting application %s: %v", appAddr, err)
			continue
		}

		if !slices.Contains(onchainApp.DelegateeGatewayAddresses, s.gatewayAddress) {
			s.logger.Warn().Msgf("GetApps: Application %s is not delegated to Gateway", onchainApp.Address)
			continue
		}

		apps = append(apps, onchainApp)
	}

	return apps, nil
}

// TODO_IMPROVE: split this function into build/sign/send/verify stages.
func (f *fullNode) SendRelay(app apptypes.Application, session sessiontypes.Session, endpoint endpoint, payload relayer.Payload) (*servicetypes.RelayResponse, error) {
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

	req, err := f.signer.SignRequest(relayRequest, app)
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
		f.accountClient,
	)
	if err != nil {
		return nil, fmt.Errorf("relay: error verifying the relay response for app %s, endpoint %s: %w", app.Address, endpoint.url, err)
	}

	return response, nil
}

// isValidUrl checks whether the provided string is a formatted as the poktroll SDK expects
// The gRPC url requires a port
func isValidUrl(urlToCheck string, needPort bool) bool {
	u, err := url.Parse(urlToCheck)
	if err != nil {
		return false
	}

	if u.Scheme == "" || u.Host == "" {
		return false
	}

	if !needPort {
		return true
	}

	_, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return false
	}

	if port == "" {
		return false
	}

	return true
}

func isValidGrpcHostPort(hostPort string) bool {
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return false
	}

	if host == "" || port == "" {
		return false
	}

	return true
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

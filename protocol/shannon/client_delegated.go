package shannon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/pokt-network/poktroll/pkg/polylog"
	apptypes "github.com/pokt-network/poktroll/x/application/types"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
	"github.com/pokt-network/shannon-sdk/client"
)

var (
	// Delegated gateway mode: could not extract app from HTTP request.
	errProtocolContextSetupGetAppFromHTTPReq = errors.New("error getting the selected app from the HTTP request")
	// Delegated gateway mode: could not fetch session for app from the full node
	errProtocolContextSetupFetchSession = errors.New("error getting a session from the full node for app")
	// Delegated gateway mode: app is not staked for the service.
	errProtocolContextSetupAppNotStaked = errors.New("app is not staked for the service")
	// Delegated gateway mode: gateway does not have delegation for the app.
	errProtocolContextSetupAppDoesNotDelegate = errors.New("gateway does not have delegation for app")
)

// - TODO(@Olshansk): Revisit the security specification & requirements for how the paying app is selected.
// - TODO_DOCUMENT(@Olshansk): Convert the Notion doc into a proper README.
// - For more details, see:
//   https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5?pvs=4#122a36edfff680eea2fbd46c7696d845

// delegatedGatewayClient implements the GatewayClient interface for delegated gateway mode.
//
// # Delegated Gateway Mode - Shannon Protocol Integration
//
//   - Represents a gateway operation mode with the following behavior:
//   - Each relay request is signed by the gateway key and sent on behalf of an app selected by the user.
//   - Users must select a specific app for each relay request (currently via HTTP request headers).
type delegatedGatewayClient struct {
	logger polylog.Logger
	*client.GatewayClient
}

// httpHeaderAppAddress is the HTTP header name for specifying the target application address.
const httpHeaderAppAddress = "X-App-Address"

// NewDelegatedGatewayClient creates a new delegatedGatewayClient instance.
func NewDelegatedGatewayClient(
	logger polylog.Logger,
	gatewayClient *client.GatewayClient,
) (*delegatedGatewayClient, error) {
	logger = logger.With("client_type", "delegated")

	return &delegatedGatewayClient{
		logger:        logger,
		GatewayClient: gatewayClient,
	}, nil
}

// GetGatewayModeActiveSessions implements GatewayClient interface.
//   - Returns the permitted session under Delegated gateway mode, for the supplied HTTP request.
//   - Gateway address and app address (specified in the HTTP header) are used to retrieve active sessions.
func (d *delegatedGatewayClient) GetGatewayModeActiveSessions(
	ctx context.Context,
	serviceID sdk.ServiceID,
	httpReq *http.Request,
) ([]sessiontypes.Session, error) {
	logger := d.logger.With("method", "GetActiveSessions")

	selectedAppAddr, err := getAppAddrFromHTTPReq(httpReq)
	if err != nil {
		err = fmt.Errorf("%w: error: %w",
			errProtocolContextSetupGetAppFromHTTPReq,
			err,
		)
		logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	logger.Debug().Msgf("fetching the app with the selected address %s.", selectedAppAddr)

	selectedSession, err := d.FullNode.GetSession(ctx, serviceID, selectedAppAddr)
	if err != nil {
		err = fmt.Errorf("%w: error: %w",
			errProtocolContextSetupFetchSession,
			err,
		)
		logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	selectedApp := selectedSession.Application

	logger.Debug().Msgf("fetched the app with the selected address %s.", selectedApp.Address)

	// Skip the session's app if it is not staked for the requested service.
	if !appIsStakedForService(serviceID, selectedApp) {
		err = fmt.Errorf("%w: app: %s",
			errProtocolContextSetupAppNotStaked,
			selectedApp.Address,
		)
		logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	if !d.gatewayHasDelegationForApp(selectedApp) {
		err = fmt.Errorf("%w: gateway: %s, app: %s",
			errProtocolContextSetupAppDoesNotDelegate,
			d.GetGatewayAddress(),
			selectedApp.Address,
		)
		logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	logger.Debug().Msgf("successfully verified the gateway has delegation for the selected app with address %s.", selectedApp.Address)

	return []sessiontypes.Session{selectedSession}, nil
}

// GetConfiguredServiceIDs is a no-op for delegated mode because
// the app address is known only at request time.
func (d *delegatedGatewayClient) GetConfiguredServiceIDs() map[sdk.ServiceID]struct{} {
	return nil
}

// gatewayHasDelegationForApp returns true if the supplied application delegates to the supplied gateway address.
func (d *delegatedGatewayClient) gatewayHasDelegationForApp(app *apptypes.Application) bool {
	return slices.Contains(app.DelegateeGatewayAddresses, d.GetGatewayAddress())
}

// appIsStakedForService returns true if the supplied application is staked for the supplied service ID.
func appIsStakedForService(serviceID sdk.ServiceID, app *apptypes.Application) bool {
	for _, svcCfg := range app.ServiceConfigs {
		if sdk.ServiceID(svcCfg.ServiceId) == serviceID {
			return true
		}
	}
	return false
}

// getAppAddrFromHTTPReq extracts the application address specified by the supplied HTTP request's headers.
func getAppAddrFromHTTPReq(httpReq *http.Request) (string, error) {
	if httpReq == nil || len(httpReq.Header) == 0 {
		return "", fmt.Errorf("getAppAddrFromHTTPReq: no HTTP headers supplied")
	}

	selectedAppAddr := httpReq.Header.Get(httpHeaderAppAddress)
	if selectedAppAddr == "" {
		return "", fmt.Errorf("getAppAddrFromHTTPReq: a target app must be supplied as HTTP header %s", httpHeaderAppAddress)
	}

	return selectedAppAddr, nil
}

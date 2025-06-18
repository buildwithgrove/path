package shannon

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"
	sessiontypes "github.com/pokt-network/poktroll/x/session/types"
	sdk "github.com/pokt-network/shannon-sdk"
	"github.com/pokt-network/shannon-sdk/client"
)

// Delegated gateway mode: could not extract app from HTTP request.
var errProtocolContextSetupGetAppFromHTTPReq = errors.New("error getting the selected app from the HTTP request")

// - TODO(@Olshansk): Revisit the security specification & requirements for how the paying app is selected.
// - TODO_DOCUMENT(@Olshansk): Convert the Notion doc into a proper README.
// - For more details, see:
//   https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5?pvs=4#122a36edfff680eea2fbd46c7696d845

// delegatedGatewayClient implements the GatewayClient interface for delegated gateway mode.
//
// It embeds the GatewayClient interface from the Shannon SDK package, which provides the
// functionality needed by the gateway package for handling service requests.
//
// # Delegated Gateway Mode - Shannon Protocol Integration
//
//   - Represents a gateway operation mode with the following behavior:
//   - Each relay request is signed by the gateway key and sent on behalf of an app selected by the user.
//   - Users must select a specific app for each relay request (currently via HTTP request headers).
type delegatedGatewayClient struct {
	logger polylog.Logger

	// Embeds the GatewayClient interface from the Shannon SDK package, which provides the
	// functionality needed by the gateway package for handling service requests.
	*client.GatewayClient
}

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

// httpHeaderAppAddress is the HTTP header name for specifying the target application address.
const httpHeaderAppAddress = "X-App-Address"

// getGatewayModeActiveSessions implements GatewayClient interface.
//   - Returns the permitted session under Delegated gateway mode, for the supplied HTTP request.
//   - App address specified in the HTTP header is used to retrieve active session for a service.
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

	return d.GetActiveSessions(ctx, serviceID, []string{selectedAppAddr})
}

// GetConfiguredServiceIDs is a no-op for delegated mode because
// the app address is known only at request time.
func (d *delegatedGatewayClient) GetConfiguredServiceIDs() map[sdk.ServiceID]struct{} {
	return nil
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

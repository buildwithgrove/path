package gateway

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/buildwithgrove/path/protocol"
)

// handleFallbackRequest processes a service request using a pre-configured fallback URL
// when no protocol-level endpoints are available for the requested service.
//
// Observations:
//   - Gateway-level: Errors are tracked for monitoring fallback reliability
//   - Protocol-level: Bypassed since fallback doesn't use protocol endpoints
//   - QoS-level: Response is provided to QoS context to allow `GetHTTPResponse`
//     method to return the response from the fallback URL to the user. Observations
//     are not provided to the QoS context since fallback requests bypass the protocol layer.
//
// Note: This is a fallback mechanism and should not be the primary request path.
// High usage of fallback URLs indicates issues with protocol endpoint availability.
func (rc *requestContext) handleFallbackRequest(payload protocol.Payload) error {
	// Construct the full fallback URL by appending the request path to the base fallback URL.
	fallbackURL := rc.fallbackURL.String()
	if payload.Path != "" {
		fallbackURL = fmt.Sprintf("%s%s", fallbackURL, payload.Path)
	}

	logger := rc.logger.With(
		"method", "handleFallbackRequest",
		"fallback_url", rc.fallbackURL.String(),
		"url_path", payload.Path,
	)

	logger.Debug().Msg("Sending fallback request")

	// Create an HTTP request using the payload data.
	// The payload.Data contains the original request body (e.g., JSON-RPC payload, REST data).
	fallbackReq, err := http.NewRequest(
		payload.Method,
		fallbackURL,
		io.NopCloser(bytes.NewReader([]byte(payload.Data))),
	)
	if err != nil {
		logger.Info().Err(err).Msg("Failed to create HTTP request for fallback URL")
		rc.updateGatewayObservations(errFallbackRequestCreationFailed)
		return errFallbackRequestCreationFailed
	}

	// TODO_IN_THIS_PR(@commoddity): add fallback HTTP client with proper configuration
	// to Gateway in order to reuse HTTP client for fallback requests.
	httpClient := http.Client{
		Timeout: time.Duration(payload.TimeoutMillisec) * time.Millisecond,
	}

	// Send the HTTP request to the fallback URL.
	// This bypasses the normal protocol layer and directly contacts the fallback endpoint.
	resp, err := httpClient.Do(fallbackReq)
	if err != nil {
		logger.Info().Err(err).Msg("Failed to send fallback request")
		rc.updateGatewayObservations(errFallbackRequestSendFailed)
		return errFallbackRequestSendFailed
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Info().Err(err).Msg("Failed to read fallback response body")
		rc.updateGatewayObservations(errFallbackResponseReadFailed)
		return errFallbackResponseReadFailed
	}

	// Update the QoS context with the fallback response so that the `GetHTTPResponse` method
	// returns the response from the fallback URL to the user.
	//
	// We can use an empty endpoint address because fallback requests do not generate
	// any QoS-level observations and so no endpoint address is used by the QoS package.
	rc.qosCtx.UpdateWithResponse(protocol.EndpointAddr(""), body)

	return nil
}

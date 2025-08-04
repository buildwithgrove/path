package gateway

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/buildwithgrove/path/protocol"
)

func (rc *requestContext) handleFallbackRequest(payload protocol.Payload) error {
	// Get the fallback URL and append the path if it exists.
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

	// TODO_IN_THIS_PR(@commoddity): add proper http client configuration.
	httpClient := http.Client{
		Timeout: time.Duration(payload.TimeoutMillisec) * time.Millisecond,
	}

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

	// An empty endpoint address is used because fallback requests are not handled by the protocol.
	//
	// Protocol and QoS-level observations are not applicable to fallback requests, so the
	// QoS package will not receive any observations from fallback requests, which is the
	// usual reason for sending the endpoint address to the `UpdateWithResponse` method.
	rc.qosCtx.UpdateWithResponse(protocol.EndpointAddr(""), body)

	return nil
}

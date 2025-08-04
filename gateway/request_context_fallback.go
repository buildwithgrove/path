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
	logger := rc.logger.With(
		"method", "handleFallbackRequest",
		"fallback_url", rc.fallbackURL.String(),
	)

	logger.Debug().Msg("Sending fallback request")

	url := rc.fallbackURL.String()
	if payload.Path != "" {
		url = fmt.Sprintf("%s%s", url, payload.Path)
	}

	fallbackReq, err := http.NewRequest(
		payload.Method,
		url,
		io.NopCloser(bytes.NewReader([]byte(payload.Data))),
	)
	if err != nil {
		logger.Info().Err(err).Str("url", url).Msg("Failed to create HTTP request for fallback URL")
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

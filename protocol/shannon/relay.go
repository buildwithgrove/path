package shannon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	servicetypes "github.com/pokt-network/poktroll/x/service/types"
)

// sendHttpRelay sends the relay request to the supplier at the given URL using an HTTP Post request.
func sendHttpRelay(
	ctx context.Context,
	supplierUrlStr string,
	relayRequest *servicetypes.RelayRequest,
) (httpRelayResponseBz []byte, err error) {
	_, err = url.Parse(supplierUrlStr)
	if err != nil {
		return nil, err
	}

	relayRequestBz, err := relayRequest.Marshal()
	if err != nil {
		return nil, err
	}

	relayHTTPRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		supplierUrlStr,
		io.NopCloser(bytes.NewReader(relayRequestBz)),
	)
	if err != nil {
		return nil, err
	}

	relayHTTPRequest.Header.Add("Content-Type", "application/json")

	var clientTimeout time.Duration
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		// Context has timeout, use a slightly longer client timeout as fallback
		remaining := time.Until(deadline)
		clientTimeout = remaining + (2 * time.Second)
	} else {
		// No context timeout, use the default keep alive time
		clientTimeout = defaultKeepAliveTime
	}

	// Create custom HTTP client with timeout
	// Ref: https://vishnubharathi.codes/blog/know-when-to-break-up-with-go-http-defaultclient/
	client := &http.Client{
		Timeout: clientTimeout,
		// TODO_IMPROVE: Allow PATH users to override default transport configs
		Transport: http.DefaultTransport,
	}

	// Send the HTTP relay request
	relayHTTPResponse, err := client.Do(relayHTTPRequest)
	if err != nil {
		return nil, err
	}
	defer relayHTTPResponse.Body.Close()

	// Read response body
	responseBody, readErr := io.ReadAll(relayHTTPResponse.Body)
	if readErr != nil {
		return nil, readErr
	}

	// Validate HTTP status code is a 2xx code
	if relayHTTPResponse.StatusCode < http.StatusOK || relayHTTPResponse.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("non-2xx status code: %d", relayHTTPResponse.StatusCode)
	}

	return responseBody, nil
}

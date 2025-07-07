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

	shannonmetrics "github.com/buildwithgrove/path/metrics/protocol/shannon"
	"github.com/buildwithgrove/path/protocol"
)

// sendHttpRelay sends the relay request to the supplier at the given URL using an HTTP Post request.
func sendHttpRelay(
	ctx context.Context,
	supplierUrlStr string,
	relayRequest *servicetypes.RelayRequest,
	timeout time.Duration,
) (relayResponseBz []byte, err error) {
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

	// TODO_IMPROVE(@commoddity): Use a custom HTTP client to:
	//  - allow configuring the defaultTransport.
	//  - allow PATH users to override default transport config.

	// Best practice in Go is to use a custom HTTP client Transport.
	// See: https://vishnubharathi.codes/blog/know-when-to-break-up-with-go-http-defaultclient/
	client := &http.Client{
		Timeout: timeout,
	}

	// Extract labels for backend service latency metrics
	serviceID := extractServiceIDFromContext(ctx)
	endpointDomain, err := shannonmetrics.ExtractDomainOrHost(supplierUrlStr)
	if err != nil {
		endpointDomain = fmt.Sprintf("error extracting domain from: %s", supplierUrlStr)
	}

	// Record backend service latency metrics
	backendStartTime := time.Now()
	relayHTTPResponse, err := client.Do(relayHTTPRequest)
	backendDuration := time.Since(backendStartTime).Seconds()

	// Prepare metrics labels
	httpStatus := "timeout"
	requestSizeBucket := categorizeRequestSize(len(relayRequestBz))

	if err != nil {
		// Record failed backend request latency
		shannonmetrics.RecordBackendServiceLatency(serviceID, endpointDomain, httpStatus, requestSizeBucket, backendDuration)
		return nil, err
	}
	defer relayHTTPResponse.Body.Close()

	// Update HTTP status for successful requests
	httpStatus = categorizeHTTPStatus(relayHTTPResponse.StatusCode)

	// Read response body
	responseBody, readErr := io.ReadAll(relayHTTPResponse.Body)
	if readErr != nil {
		// Record backend latency even for read errors
		shannonmetrics.RecordBackendServiceLatency(serviceID, endpointDomain, httpStatus, requestSizeBucket, backendDuration)
		return nil, readErr
	}

	// Record successful backend service latency
	shannonmetrics.RecordBackendServiceLatency(serviceID, endpointDomain, httpStatus, requestSizeBucket, backendDuration)

	return responseBody, nil
}

// extractServiceIDFromContext extracts service ID from context (simplified version)
// TODO_IMPROVE: Pass service ID explicitly through function parameters instead of context
func extractServiceIDFromContext(ctx context.Context) protocol.ServiceID {
	// This is a simplified implementation. In practice, you might want to
	// pass the service ID more explicitly through the function parameters
	if serviceID := ctx.Value("service_id"); serviceID != nil {
		if str, ok := serviceID.(string); ok {
			return protocol.ServiceID(str)
		}
	}
	return protocol.ServiceID("unknown")
}

// categorizeRequestSize buckets request size for metrics
func categorizeRequestSize(size int) string {
	switch {
	case size < 1024: // < 1KB
		return "small"
	case size < 10240: // < 10KB
		return "medium"
	default: // >= 10KB
		return "large"
	}
}

// categorizeHTTPStatus converts HTTP status codes to metric categories
func categorizeHTTPStatus(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "2xx"
	case statusCode >= 400 && statusCode < 500:
		return "4xx"
	case statusCode >= 500:
		return "5xx"
	default:
		return "other"
	}
}

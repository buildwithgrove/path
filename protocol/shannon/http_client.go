package shannon

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
)

// httpClientWithDebugMetrics provides HTTP client functionality with debugging functionality.
// It includes things like:
// - Built-in request debugging
// - Metrics collection
// - Detailed logging
// - Timeout debugging
// - Connection issue visibility
type httpClientWithDebugMetrics struct {
	httpClient *http.Client

	// Atomic counters for monitoring
	activeRequests   atomic.Uint64
	totalRequests    atomic.Uint64
	timeoutErrors    atomic.Uint64
	connectionErrors atomic.Uint64
}

// requestMetrics holds detailed timing and status information for a single HTTP request
type requestMetrics struct {
	startTime      time.Time
	dnsLookupTime  time.Duration
	connectTime    time.Duration
	tlsTime        time.Duration
	firstByteTime  time.Duration
	totalTime      time.Duration
	statusCode     int
	error          error
	contextTimeout time.Duration
	goroutineCount int
	url            string
}

// newDefaultHTTPClientWithDebugMetrics creates a new HTTP client with optimized transport settings
// and built-in request debugging capabilities using default configuration.
// TODO_TECHDEBT(@adshmh): Make HTTP client settings configurable
func newDefaultHTTPClientWithDebugMetrics() *httpClientWithDebugMetrics {
	// Configure transport with optimized settings for high-concurrency usage
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second, // Connection establishment timeout
			KeepAlive: 30 * time.Second, // Keep-alive probe interval
		}).DialContext,

		// Connection pool settings
		MaxIdleConns:        100,              // Total idle connections across all hosts
		MaxIdleConnsPerHost: 10,               // Idle connections per host
		MaxConnsPerHost:     50,               // Max concurrent connections per host
		IdleConnTimeout:     90 * time.Second, // How long idle connections stay open

		// Timeout settings
		TLSHandshakeTimeout:   10 * time.Second, // TLS handshake timeout
		ResponseHeaderTimeout: 30 * time.Second, // Time to wait for response headers

		// Performance settings
		DisableKeepAlives:  false, // Enable connection reuse
		DisableCompression: false, // Enable gzip compression
	}

	// Create HTTP client with large timeout as fallback
	// Individual requests will use context deadlines for actual timeout control
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second, // Large fallback timeout (1 minute)
	}

	return &httpClientWithDebugMetrics{
		httpClient: httpClient,
	}
}

// SendHTTPRelay sends an HTTP POST request with the relay data to the specified URL.
// Uses the provided context for timeout and cancellation control.
// Logs detailed metrics and debugging information on failure for debugging.
func (h *httpClientWithDebugMetrics) SendHTTPRelay(
	ctx context.Context,
	logger polylog.Logger,
	endpointURL string,
	relayRequest *servicetypes.RelayRequest,
	headers map[string]string,
) ([]byte, error) {
	// Set up debugging context and logging function
	debugCtx, recordRequest := h.setupRequestDebugging(ctx, logger, endpointURL)

	var requestErr error
	defer func() {
		recordRequest(requestErr)
	}()

	// Validate URL format
	_, err := url.Parse(endpointURL)
	if err != nil {
		requestErr = fmt.Errorf("SHOULD NEVER HAPPEN: invalid URL: %w", err)
		return nil, requestErr
	}

	// Marshal relay request to bytes
	relayRequestBz, err := relayRequest.Marshal()
	if err != nil {
		requestErr = fmt.Errorf("SHOULD NEVER HAPPEN: failed to marshal relay request: %w", err)
		return nil, requestErr
	}

	req, err := http.NewRequestWithContext(
		debugCtx,
		http.MethodPost,
		endpointURL,
		bytes.NewReader(relayRequestBz),
	)
	if err != nil {
		requestErr = fmt.Errorf("failed to create HTTP request: %w", err)
		return nil, requestErr
	}

	// TOCO_TECHDEBT(@adshmh): Content-Type HTTP header should be set by the QoS.
	//
	// Set HTTP headers.
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute HTTP request
	resp, err := h.httpClient.Do(req)
	if err != nil {
		requestErr = h.categorizeError(debugCtx, err)
		return nil, requestErr
	}
	defer resp.Body.Close()

	// Read and validate response
	responseBody, err := h.readAndValidateResponse(resp)
	if err != nil {
		requestErr = err
		return nil, requestErr
	}

	return responseBody, nil
}

// setupRequestDebugging initializes request metrics, HTTP debugging context, and atomic counters.
// Returns the debug context and a cleanup function that accepts an error parameter.
func (h *httpClientWithDebugMetrics) setupRequestDebugging(
	ctx context.Context,
	logger polylog.Logger,
	endpointURL string,
) (context.Context, func(error)) {
	// Update atomic counters
	h.activeRequests.Add(1)
	h.totalRequests.Add(1)

	startTime := time.Now()

	// Initialize metrics collection
	metrics := &requestMetrics{
		startTime:      startTime,
		goroutineCount: runtime.NumGoroutine(),
		url:            endpointURL,
	}

	// Capture context timeout for logging
	if deadline, ok := ctx.Deadline(); ok {
		metrics.contextTimeout = time.Until(deadline)
	}

	// Create HTTP trace and add to context
	trace := createHTTPTrace(metrics)
	debugCtx := httptrace.WithClientTrace(ctx, trace)

	// Return recorder function that logs request details.
	requestRecorder := func(err error) {
		h.activeRequests.Add(^uint64(0)) // Atomic decrement
		metrics.totalTime = time.Since(metrics.startTime)
		metrics.error = err
		if err != nil {
			h.logRequestMetrics(logger, *metrics)
		}
	}

	return debugCtx, requestRecorder
}

// categorizeError categorizes HTTP client errors and updates counters for monitoring
func (h *httpClientWithDebugMetrics) categorizeError(ctx context.Context, err error) error {
	if ctx.Err() == context.DeadlineExceeded {
		h.timeoutErrors.Add(1)
		return fmt.Errorf("request timeout: %w", err)
	} else {
		h.connectionErrors.Add(1)
		return fmt.Errorf("connection error: %w", err)
	}
}

// readAndValidateResponse reads the response body and validates the HTTP status code
func (h *httpClientWithDebugMetrics) readAndValidateResponse(resp *http.Response) ([]byte, error) {
	// Read response body with size protection
	const maxResponseSize = 100 * 1024 * 1024 // 100MB limit
	limitedReader := io.LimitReader(resp.Body, maxResponseSize)
	responseBody, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Validate HTTP status code
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%w: %d", errRelayEndpointHTTPError, resp.StatusCode)
	}

	return responseBody, nil
}

// createHTTPTrace creates an HTTP trace that captures timing metrics
// for each phase of the HTTP request lifecycle.
func createHTTPTrace(metrics *requestMetrics) *httptrace.ClientTrace {
	var dnsStart, connectStart, tlsStart time.Time

	return &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			if !dnsStart.IsZero() {
				metrics.dnsLookupTime = time.Since(dnsStart)
			}
		},
		ConnectStart: func(network, addr string) {
			connectStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			if !connectStart.IsZero() {
				metrics.connectTime = time.Since(connectStart)
			}
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if !tlsStart.IsZero() {
				metrics.tlsTime = time.Since(tlsStart)
			}
		},
		GotFirstResponseByte: func() {
			metrics.firstByteTime = time.Since(metrics.startTime)
		},
	}
}

// logRequestMetrics logs comprehensive request metrics for debugging failed requests.
// Only called when a request fails to avoid verbose logging on successful requests.
func (h *httpClientWithDebugMetrics) logRequestMetrics(logger polylog.Logger, metrics requestMetrics) {
	// Log detailed failure metrics using the provided structured logger
	logger.With(
		"http_client_debug_url", metrics.url,
		"http_client_debug_dns_lookup_ms", metrics.dnsLookupTime.Milliseconds(),
		"http_client_debug_connect_ms", metrics.connectTime.Milliseconds(),
		"http_client_debug_tls_ms", metrics.tlsTime.Milliseconds(),
		"http_client_debug_first_byte_ms", metrics.firstByteTime.Milliseconds(),
		"http_client_debug_total_ms", metrics.totalTime.Milliseconds(),
		"http_client_debug_status_code", metrics.statusCode,
		"http_client_debug_timeout_ms", metrics.contextTimeout.Milliseconds(),
		"http_client_debug_goroutines", metrics.goroutineCount,
		"http_client_debug_active_requests", h.activeRequests.Load(),
		"http_client_debug_total_requests", h.totalRequests.Load(),
		"http_client_debug_timeout_errors", h.timeoutErrors.Load(),
		"http_client_debug_connection_errors", h.connectionErrors.Load(),
	).Error().Err(metrics.error).Msg("HTTP request failed - detailed timing breakdown")
}

// TODO_TECHDEBT(@adshmh): Add graceful shutdown support.

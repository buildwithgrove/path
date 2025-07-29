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

// httpClientWithTracing provides HTTP client functionality with built-in request tracing,
// metrics collection, and detailed logging for debugging timeout and connection issues.
type httpClientWithTracing struct {
	httpClient *http.Client

	// Atomic counters for monitoring
	activeRequests   int64
	totalRequests    int64
	timeoutErrors    int64
	connectionErrors int64
}

// requestMetrics holds detailed timing and status information for a single HTTP request
type requestMetrics struct {
	StartTime      time.Time
	DNSLookupTime  time.Duration
	ConnectTime    time.Duration
	TLSTime        time.Duration
	FirstByteTime  time.Duration
	TotalTime      time.Duration
	StatusCode     int
	Error          error
	ContextTimeout time.Duration
	GoroutineCount int
	URL            string
}

// newHTTPClientWithDefaultTracing creates a new HTTP client with optimized transport settings
// and built-in request tracing capabilities using default configuration.
// TODO_TECHDEBT(@adshmh): Make HTTP client settings configurable
func newHTTPClientWithDefaultTracing() *httpClientWithTracing {
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

	return &httpClientWithTracing{
		httpClient: httpClient,
	}
}

// SendHTTPRelay sends an HTTP POST request with the relay data to the specified URL.
// Uses the provided context for timeout and cancellation control.
// Logs detailed metrics and tracing information on failure for debugging.
func (h *httpClientWithTracing) SendHTTPRelay(
	ctx context.Context,
	logger polylog.Logger,
	endpointURL string,
	relayRequest *servicetypes.RelayRequest,
	headers map[string]string,
) ([]byte, error) {
	// Set up tracing context and logging function
	tracedCtx, recordRequest := h.setupRequestTracing(ctx, logger, endpointURL)

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
		tracedCtx,
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
		requestErr = h.categorizeError(tracedCtx, err)
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

// setupRequestTracing initializes request metrics, HTTP tracing context, and atomic counters.
// Returns the traced context and a cleanup function that accepts an error parameter.
func (h *httpClientWithTracing) setupRequestTracing(
	ctx context.Context,
	logger polylog.Logger,
	endpointURL string,
) (context.Context, func(error)) {
	// Update atomic counters
	atomic.AddInt64(&h.activeRequests, 1)
	atomic.AddInt64(&h.totalRequests, 1)

	startTime := time.Now()

	// Initialize metrics collection
	metrics := &requestMetrics{
		StartTime:      startTime,
		GoroutineCount: runtime.NumGoroutine(),
		URL:            endpointURL,
	}

	// Capture context timeout for logging
	if deadline, ok := ctx.Deadline(); ok {
		metrics.ContextTimeout = time.Until(deadline)
	}

	// Create HTTP trace and add to context
	trace := createHTTPTrace(metrics)
	tracedCtx := httptrace.WithClientTrace(ctx, trace)

	// Return recorder function that logs request details.
	requestRecorder := func(err error) {
		atomic.AddInt64(&h.activeRequests, -1)
		metrics.TotalTime = time.Since(metrics.StartTime)
		metrics.Error = err
		if err != nil {
			h.logRequestMetrics(logger, *metrics)
		}
	}

	return tracedCtx, requestRecorder
}

// categorizeError categorizes HTTP client errors and updates counters for monitoring
func (h *httpClientWithTracing) categorizeError(ctx context.Context, err error) error {
	if ctx.Err() == context.DeadlineExceeded {
		atomic.AddInt64(&h.timeoutErrors, 1)
		return fmt.Errorf("request timeout: %w", err)
	} else {
		atomic.AddInt64(&h.connectionErrors, 1)
		return fmt.Errorf("connection error: %w", err)
	}
}

// readAndValidateResponse reads the response body and validates the HTTP status code
func (h *httpClientWithTracing) readAndValidateResponse(resp *http.Response) ([]byte, error) {
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
				metrics.DNSLookupTime = time.Since(dnsStart)
			}
		},
		ConnectStart: func(network, addr string) {
			connectStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			if !connectStart.IsZero() {
				metrics.ConnectTime = time.Since(connectStart)
			}
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if !tlsStart.IsZero() {
				metrics.TLSTime = time.Since(tlsStart)
			}
		},
		GotFirstResponseByte: func() {
			metrics.FirstByteTime = time.Since(metrics.StartTime)
		},
	}
}

// logRequestMetrics logs comprehensive request metrics for debugging failed requests.
// Only called when a request fails to avoid verbose logging on successful requests.
func (h *httpClientWithTracing) logRequestMetrics(logger polylog.Logger, metrics requestMetrics) {
	// Log detailed failure metrics using the provided structured logger
	logger.With(
		"url", metrics.URL,
		"dns_lookup_ms", metrics.DNSLookupTime.Milliseconds(),
		"connect_ms", metrics.ConnectTime.Milliseconds(),
		"tls_ms", metrics.TLSTime.Milliseconds(),
		"first_byte_ms", metrics.FirstByteTime.Milliseconds(),
		"total_ms", metrics.TotalTime.Milliseconds(),
		"status_code", metrics.StatusCode,
		"timeout_ms", metrics.ContextTimeout.Milliseconds(),
		"goroutines", metrics.GoroutineCount,
		"active_requests", atomic.LoadInt64(&h.activeRequests),
		"total_requests", atomic.LoadInt64(&h.totalRequests),
		"timeout_errors", atomic.LoadInt64(&h.timeoutErrors),
		"connection_errors", atomic.LoadInt64(&h.connectionErrors),
	).Error().Err(metrics.Error).Msg("HTTP request failed - detailed timing breakdown")
}

// Close gracefully shuts down the HTTP client and logs final statistics
func (h *httpClientWithTracing) Close() {
	// Log basic shutdown info
	fmt.Printf("HTTP_CLIENT_SHUTDOWN: active_requests=%d total_requests=%d timeout_errors=%d connection_errors=%d\n",
		atomic.LoadInt64(&h.activeRequests),
		atomic.LoadInt64(&h.totalRequests),
		atomic.LoadInt64(&h.timeoutErrors),
		atomic.LoadInt64(&h.connectionErrors),
	)

	// Close idle connections
	h.httpClient.CloseIdleConnections()
}

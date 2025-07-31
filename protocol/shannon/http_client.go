package shannon

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"
	servicetypes "github.com/pokt-network/poktroll/x/service/types"
)

// Maximum length of an HTTP response's body.
const maxResponseSize = 100 * 1024 * 1024 // 100MB limit

// Fast-fail timeout for detecting blocked request body writes
// If headers are written but request body isn't written within this time, fail fast
const requestBodyWriteTimeout = 1 * time.Second
const recentHangingEndpointTimeout = 1 * time.Minute

// httpClientWithDebugMetrics provides HTTP client functionality with embedded tracking of debug metrics.
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

	// Track hanging endpoints for fast-fail behavior
	hangingEndpoints map[string]time.Time // URL -> last detection time
	hangingMutex     sync.RWMutex         // Protects hangingEndpoints map
}

// httpRequestMetrics holds detailed timing and status information for a single HTTP request
type httpRequestMetrics struct {
	startTime      time.Time
	url            string
	contextTimeout time.Duration
	goroutineCount int

	// DNS Resolution
	dnsLookupTime time.Duration

	// Connection Establishment
	connectTime      time.Duration
	connectionReused bool
	remoteAddr       string
	localAddr        string

	// TLS Handshake
	tlsTime time.Duration

	// Connection Acquisition (from pool or new)
	getConnTime time.Duration

	// Request Writing
	wroteHeadersTime time.Duration
	wroteRequestTime time.Duration

	// Response Waiting
	firstByteTime time.Duration

	// Overall
	totalTime  time.Duration
	statusCode int
	error      error
}

// newDefaultHTTPClientWithDebugMetrics creates a new HTTP client with:
// - Transport settings configured for high-concurrency usage
// - Built in request debugging capabilities and metrics tracking
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
		ResponseHeaderTimeout: 70 * time.Second, // Time to wait for response headers

		// Performance settings
		DisableKeepAlives:  false, // Enable connection reuse
		DisableCompression: false, // Enable gzip compression
	}

	// Create HTTP client with large timeout as fallback
	// Individual requests will use context deadlines for actual timeout control
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   80 * time.Second, // Large fallback timeout (80 seconds)
	}

	return &httpClientWithDebugMetrics{
		httpClient:       httpClient,
		hangingEndpoints: make(map[string]time.Time),
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
	// Fast-fail check: if this endpoint is known to be hanging, fail immediately
	h.hangingMutex.RLock()
	if lastDetection, isHanging := h.hangingEndpoints[endpointURL]; isHanging {
		// Only fast-fail if the detection was recent (within last 5 minutes)
		if time.Since(lastDetection) < recentHangingEndpointTimeout {
			h.hangingMutex.RUnlock()
			return nil, fmt.Errorf("fast-fail: endpoint %s is known hanging (blocks request body writes)", endpointURL)
		}
		// Detection is old, remove it and continue
		h.hangingMutex.RUnlock()
		h.hangingMutex.Lock()
		delete(h.hangingEndpoints, endpointURL)
	}
	h.hangingMutex.Unlock()

	// Set up debugging context and logging function
	debugCtx, requestRecorder := h.setupRequestDebugging(ctx, logger, endpointURL)

	var requestErr error
	defer func() {
		requestRecorder(requestErr)
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

	bodyReader := wrapBodyWithFailFast(relayRequestBz, requestBodyWriteTimeout)

	req, err := http.NewRequestWithContext(
		debugCtx,
		http.MethodPost,
		endpointURL,
		bodyReader,
	)
	if err != nil {
		requestErr = fmt.Errorf("failed to create HTTP request: %w", err)
		return nil, requestErr
	}

	// TODO_TECHDEBT(@adshmh): Content-Type HTTP header should be set by the QoS.
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
	metrics := &httpRequestMetrics{
		startTime:      startTime,
		goroutineCount: runtime.NumGoroutine(),
		url:            endpointURL,
	}

	// Capture context timeout for logging
	if deadline, ok := ctx.Deadline(); ok {
		metrics.contextTimeout = time.Until(deadline)
	}

	// Create HTTP trace and add to context
	trace := createDetailedHTTPTrace(metrics)
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

// createDetailedHTTPTrace creates comprehensive HTTP tracing using the httptrace library:
// https://pkg.go.dev/net/http/httptrace
// Captures granular timing for every phase of the HTTP request lifecycle to identify bottlenecks.
func createDetailedHTTPTrace(metrics *httpRequestMetrics) *httptrace.ClientTrace {
	var (
		dnsStart, connectStart, tlsStart time.Time
		getConnStart, wroteRequestStart  time.Time
		waitingForResponseStart          time.Time
	)

	return &httptrace.ClientTrace{
		// DNS Resolution Phase
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			if !dnsStart.IsZero() {
				metrics.dnsLookupTime = time.Since(dnsStart)
			}
		},

		// Connection Establishment Phase
		ConnectStart: func(network, addr string) {
			connectStart = time.Now()
			metrics.remoteAddr = addr
		},
		ConnectDone: func(network, addr string, err error) {
			if !connectStart.IsZero() {
				metrics.connectTime = time.Since(connectStart)
			}
		},

		// TLS Handshake Phase
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			if !tlsStart.IsZero() {
				metrics.tlsTime = time.Since(tlsStart)
			}
		},

		// Connection Acquisition Phase
		// Tracks potential connection pool exhaustion.
		GetConn: func(hostPort string) {
			getConnStart = time.Now()
		},
		GotConn: func(info httptrace.GotConnInfo) {
			if !getConnStart.IsZero() {
				metrics.getConnTime = time.Since(getConnStart)
			}
			metrics.connectionReused = info.Reused
			if info.Conn != nil {
				metrics.localAddr = info.Conn.LocalAddr().String()
			}
		},

		// Request Writing Phase
		// Tracks potential write delays.
		WroteHeaders: func() {
			// Headers written successfully, time from connection acquisition to headers completion
			if !getConnStart.IsZero() {
				metrics.wroteHeadersTime = time.Since(getConnStart)
			}
			// Start timing request body writing
			wroteRequestStart = time.Now()
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			// Entire request (headers + body) written successfully
			if !wroteRequestStart.IsZero() {
				metrics.wroteRequestTime = time.Since(wroteRequestStart)
			}
			// Now waiting for server response
			waitingForResponseStart = time.Now()
		},

		// Response Reading Phase
		GotFirstResponseByte: func() {
			if !waitingForResponseStart.IsZero() {
				metrics.firstByteTime = time.Since(waitingForResponseStart)
			}
		},
	}
}

// logRequestMetrics logs comprehensive request metrics for debugging failed requests.
// Only called when a request fails to avoid verbose logging on successful requests.
func (h *httpClientWithDebugMetrics) logRequestMetrics(logger polylog.Logger, metrics httpRequestMetrics) {
	// Calculate derived timings for easier analysis
	connectionEstablishmentTime := metrics.dnsLookupTime + metrics.connectTime + metrics.tlsTime
	requestTransmissionTime := metrics.wroteHeadersTime + metrics.wroteRequestTime

	// Detect hanging network behavior pattern:
	// - Headers were written successfully (wroteHeadersTime > 0)
	// - But request body write failed or took 0ms (wroteRequestTime == 0)
	// - Request timed out near the full timeout duration
	isHangingPattern := metrics.wroteHeadersTime > 0 &&
		metrics.wroteRequestTime == 0 &&
		metrics.totalTime > 55*time.Second && // Close to 60s timeout
		metrics.totalTime < 62*time.Second

	// Log detailed failure metrics using the provided structured logger
	logger.With(
		// Request identification
		"http_client_debug_url", metrics.url,
		"http_client_debug_total_ms", metrics.totalTime.Milliseconds(),
		"http_client_debug_timeout_ms", metrics.contextTimeout.Milliseconds(),
		"http_client_debug_status_code", metrics.statusCode,

		// Phase 1: DNS Resolution
		"http_client_debug_dns_lookup_ms", metrics.dnsLookupTime.Milliseconds(),

		// Phase 2: Connection Management
		"http_client_debug_get_conn_ms", metrics.getConnTime.Milliseconds(), // Time to get connection from pool
		"http_client_debug_connection_reused", metrics.connectionReused, // Was connection reused?
		"http_client_debug_connect_ms", metrics.connectTime.Milliseconds(), // TCP connection time (if new)
		"http_client_debug_tls_ms", metrics.tlsTime.Milliseconds(), // TLS handshake time (if new)
		"http_client_debug_connection_establishment_ms", connectionEstablishmentTime.Milliseconds(), // Total setup time

		// Phase 3: Request Transmission
		"http_client_debug_wrote_headers_ms", metrics.wroteHeadersTime.Milliseconds(), // Time to write headers
		"http_client_debug_wrote_request_ms", metrics.wroteRequestTime.Milliseconds(), // Time to write body
		"http_client_debug_request_transmission_ms", requestTransmissionTime.Milliseconds(), // Total write time

		// Phase 4: Response Waiting
		"http_client_debug_first_byte_ms", metrics.firstByteTime.Milliseconds(), // Time waiting for server response

		// Connection details
		"http_client_debug_remote_addr", metrics.remoteAddr,
		"http_client_debug_local_addr", metrics.localAddr,

		// System state
		"http_client_debug_goroutines", metrics.goroutineCount,
		"http_client_debug_active_requests", h.activeRequests.Load(),
		"http_client_debug_total_requests", h.totalRequests.Load(),
		"http_client_debug_timeout_errors", h.timeoutErrors.Load(),
		"http_client_debug_connection_errors", h.connectionErrors.Load(),

		// Hanging behavior detection
		"http_client_debug_hanging_pattern", isHangingPattern,
	).Error().Err(metrics.error).Msg(func() string {
		if isHangingPattern {
			return "HTTP request failed - HANGING NETWORK DETECTED: Headers accepted but request body write blocked"
		}
		return "HTTP request failed - detailed phase breakdown for timeout debugging"
	}())

	// If hanging pattern detected, mark this endpoint for fast-fail
	if isHangingPattern {
		h.hangingMutex.Lock()
		h.hangingEndpoints[metrics.url] = time.Now()
		h.hangingMutex.Unlock()
		logger.Warn().
			Str("endpoint_url", metrics.url).
			Msg("Endpoint marked as hanging - future requests will fast-fail for 5 minutes")
	}
}

// wrapBodyWithFailFast wraps a []byte into an io.ReadCloser using io.Pipe,
// and fails fast if the write does not start within `writeTimeout`.
func wrapBodyWithFailFast(body []byte, writeTimeout time.Duration) io.ReadCloser {
	pr, pw := io.Pipe()

	// Signal to control when the write starts
	writeStarted := make(chan struct{})

	go func() {
		defer pw.Close()
		select {
		case <-writeStarted:
			_, err := pw.Write(body)
			if err != nil {
				_ = pw.CloseWithError(fmt.Errorf("body write error: %w", err))
			}
		case <-time.After(writeTimeout):
			_ = pw.CloseWithError(fmt.Errorf("fail-fast: request body write did not start within %s", writeTimeout))
		}
	}()

	close(writeStarted) // trigger the write immediately
	return pr
}

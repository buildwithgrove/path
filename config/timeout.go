package config

import (
	"time"
)

const (
	// HTTP server's timeout values.
	defaultHTTPServerReadTimeout = 5_000 * time.Millisecond
	defaultHTTPServerIdleTimeout = 120_000 * time.Millisecond

	// HTTP request handler's WriteTimeout.
	// https://pkg.go.dev/net/http#Server
	defaultHTTPServerWriteTimeout = 20_000 * time.Millisecond
)

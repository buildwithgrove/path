package config

import (
	"time"
)

const (
	// Used for HTTP server's timeout values.
	// https://pkg.go.dev/net/http#Server
	defaultReadTimeout  = 5_000 * time.Millisecond
	defaultWriteTimeout = 20_000 * time.Millisecond
	defaultIdleTimeout  = 120_000 * time.Millisecond

	// Max time allowed for request handling operations.
	// Calculated based on `defaultWriteTimeout` above, which sets HTTP handler's WriteTimeout.
	RequestProcessingTimeout = defaultWriteTimeout - 5_000*time.Millisecond
)

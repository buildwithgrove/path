package framework

import (
	"fmt"
	"time"
)

// TODO_FUTURE(@adshmh): Support one or both of the following:
// - ake these sanction durations/types configurable through service config,
const (
	// Default sanction duration for empty responses
	DefaultEmptyResponseSanctionDuration = 5 * time.Minute

	// Default sanction duration for parse errors
	DefaultParseErrorSanctionDuration = 5 * time.Minute

	// Default sanction duration for no responses
	DefaultNoResponseSanctionDuration = 5 * time.Minute
)

// applySanctionForNoResponse applies the default sanction for when no endpoint responded.
// This can occur due to protocol-level failures: e.g. the selected endpoint was temporarily unavailable.
// This is not the same as empty responses (where an endpoint responded with empty data).
func applySanctionForNoResponse(ctx *ResultContext) *ResultData {
	return ctx.SanctionEndpoint(
		"No endpoint responses received",
		"Protocol error",
		DefaultNoResponseSanctionDuration,
	)
}

// applySanctionForEmptyResponse applies the default sanction for empty responses.
func applySanctionForEmptyResponse(ctx *ResultContext) *ResultData {
	return ctx.SanctionEndpoint(
		"Empty response from endpoint",
		"Empty response",
		DefaultEmptyResponseSanctionDuration,
	)
}

// applySanctionForUnmarshalingError applies the default sanction for parse errors.
func applySanctionForUnmarshalingError(ctx *ResultContext, err error) *ResultData {
	return ctx.SanctionEndpoint(
		fmt.Sprintf("Failed to parse endpoint response: %v", err),
		"Parse error",
		DefaultParseErrorSanctionDuration,
	)
}

// TODO_FUTURE: Add capability to override default sanctions and/or make them configurable
// through service configuration or dynamic policy updates.

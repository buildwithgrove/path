package gateway

import (
	"context"
	"net/http"
)

// RequestResponseObserver defines the interface for reporting all the details
// regarding a request and its corresponding response and set of events
// to any interested entity.
// Examples of observers includes:
// - A request rate-limiter: to update users' served requests.
// - The QoS system: to update endpoints' latencies, endpoints' success rates, etc.
type RequestResponseObserver interface {
	// The Context input is expected to have component-specific data items
	// attached to it, which is to be retreived by calling the correct methods
	// on the corresponding package.
	ObserveReqRes(context.Context, *http.Request, HTTPResponse)
}

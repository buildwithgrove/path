package gateway

import "net/http"

// serviceRequestType represents the type of service request.
type serviceRequestType int

const (
	// httpServiceRequest represents a standard HTTP service request.
	httpServiceRequest serviceRequestType = iota
	// websocketServiceRequest represents a WebSocket connection request.
	websocketServiceRequest
	// Future request types can be added here.
)

// determineServiceRequestType checks the incoming HTTP request and returns the appropriate serviceRequestType.
func determineServiceRequestType(httpReq *http.Request) serviceRequestType {
	switch {
	case isWebsocketRequest(httpReq):
		return websocketServiceRequest
	default:
		return httpServiceRequest
	}
}

// isWebsocketRequest checks if the incoming HTTP request is a WebSocket connection request.
func isWebsocketRequest(httpReq *http.Request) bool {
	upgradeHeader := httpReq.Header.Get("Upgrade")
	connectionHeader := httpReq.Header.Get("Connection")

	return http.CanonicalHeaderKey(upgradeHeader) == "Websocket" &&
		http.CanonicalHeaderKey(connectionHeader) == "Upgrade"
}

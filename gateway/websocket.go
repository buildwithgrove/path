package gateway

import "github.com/buildwithgrove/path/observation"

// WebsocketsBridge routes data between an Endpoint and a Client.
// One bridge represents a single WebSocket connection between a Client and a WebSocket Endpoint.
//
// Full data flow:
//
//	Client <--- clientConn ---> PATH Bridge <--- endpointConn ---> Relay Miner Bridge <------> Endpoint
type WebsocketsBridge interface {
	// StartAsync starts the bridge and handles the data flow between the Client and the Endpoint.
	// It is called by the Gateway when a new WebSocket connection is established.
	//
	// IMPORTANT: StartAsync should always be run in a goroutine to avoid blocking the main thread.
	StartAsync(*observation.GatewayObservations, RequestResponseReporter)
}

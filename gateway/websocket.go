package gateway

// WebsocketsBridge routes data between an Endpoint and a Client.
// One bridge represents a single WebSocket connection between a Client and a WebSocket Endpoint.
//
// Full data flow:
//
//	Client <--- clientConn ---> PATH Bridge <--- endpointConn ---> Relay Miner Bridge <------> Endpoint
type WebsocketsBridge interface {
	// Start starts the bridge and handles the data flow between the Client and the Endpoint.
	// It is called by the Gateway when a new WebSocket connection is established.
	//
	// IMPORTANT: Start should always be run in a goroutine to avoid blocking the main thread.
	Start()
}

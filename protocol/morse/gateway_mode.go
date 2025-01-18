package morse

import (
	"github.com/buildwithgrove/path/protocol"
)

// SupportedGatewayModes returns the list of Gateway Modes supported by the Morse protocol.
// Implements the gateway.Protocol interface.
func (p *Protocol) SupportedGatewayModes() []protocol.GatewayMode {
	return []protocol.GatewayMode{
		protocol.GatewayModeCentralized,
	}
}

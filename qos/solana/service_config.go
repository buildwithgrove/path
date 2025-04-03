package solana

import "github.com/buildwithgrove/path/protocol"

// QoSType is the QoS type for the Solana blockchain.
const QoSType = "solana"

type ServiceConfig struct {
	ServiceID protocol.ServiceID
}

func (c ServiceConfig) GetServiceID() protocol.ServiceID {
	return c.ServiceID
}

func (c ServiceConfig) GetServiceQoSType() string {
	return QoSType
}

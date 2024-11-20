package protocol

// TODO_MVP(@adshmh): add a README based on the following notion doc:
// https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5
//
// GatewayMode represents the operation mode of the gateway that is using a relaying protcol for serving user requests.
// It is defined in the `protocol` package for:
// a. Consistency: protocol package defines all key concepts related to a relaying protocol.
// b. Clarity: make it clear that is a protocol-level concern to be addressed by all protocol implementations.
//
// As of now, the GatewayMode can be one of the following (see the comment attached to each mode for details).
// 1. Centralized
// 2. Delegated
// 3. Permissionless
type GatewayMode string

const (
	GatewayModeCentralized    = "centralized"
	GatewayModeDelegated      = "delegated"
	GatewayModePermissionless = "permissionless"
)

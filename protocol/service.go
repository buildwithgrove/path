package protocol

// ServiceID represents a unique onchain ID for a service.
// It is defined in the `protocol` package and not the `qos` package because `qos` is intended to handle off-chain details,
// while the `protocol` package defines onchain specs.
// See the discussion here for more details:
// https://github.com/buildwithgrove/path/pull/767#discussion_r1722001685
type ServiceID string

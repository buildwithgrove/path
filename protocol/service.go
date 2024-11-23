package protocol

// ServiceID represents a unique onchain ID for a service.
// It is defined in the `protocol` package and not the `qos` package because:
// A. Protocols (i.e. Morse and Shannon) define and maintain onchain entities, including service IDs.
// B. The `qos` package handles offchain specs & details.
// C. This allows the `gateway` package to map multiple Service IDs to a single qos implementation, e.g. all EVM blockchain services can be handled by `qos/evm`.
type ServiceID string

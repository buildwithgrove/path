// package observation defines all the structures used to communicate all aspects of an observation by each component of PATH.
// This includes, but is not limited to observations of:
// 1. HTTP service requests
// 2. Protocol-level observations: e.g. latency of an endpoint.
// 3. QoS-level observations: e.g. Solana blockchain Epoch reported by an endpoint.
//
// All the struct are generated using protobuf compiler, from .proto files under proto directory.
package observation

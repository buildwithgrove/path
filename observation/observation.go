// package observation defines all the structures used to communicate all aspects of an observation by each component of PATH.
//
// This includes, but is not limited to observations of:
// 1. HTTP service requests; e.g. the length of the HTTP request's payload.
// 2. Protocol-level observations; e.g. an endpoint's indicating it is maxed-out (i.e. over-serviced) for an app (i.e. an onchain staked application).
// 3. QoS-level observations; e.g. Solana blockchain Epoch reported by an endpoint.
//
// Recall that an endpoint is a URL of a particular onchain staked node/supplier providing service.
package observation

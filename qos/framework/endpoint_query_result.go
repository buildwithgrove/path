package framework

import (
	"fmt"
	"time"

	"github.com/buildwithgrove/path/protocol"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// TODO_IN_THIS_PR: make all the fields private, and provide Concurrency-safe methods to access Int and String values.
// This will allow the Endpoint struct to return the EndpointQueryResult struct as a whole, and simplify the client code.
// e.g.:
// Instead of:
//     - endpoint.GetQueryResultIntValue("getEpochInfo", "epoch")
//     - endpoint.GetQueryResultIntValue("getEpochInfo", "blockHeight")
// We can write:
//     - epochInfoResult := endpoint.GetQueryResult("getEpochInfo")
//     - epoch := epochInfoResult.GetIntValue("epoch")
//     - blockHeight := epochInfoResult.GetIntValue("blockHeight")

// TODO_IMPROVE(@adshmh): Enhance EndpointQueryResult to support data types commonly stored for endpoints.
//
// EndpointQueryResult captures data extracted from an endpoint query.
// - Stores one or more string/integer values.
// - Contains error/sanction information on endpoint error.
type EndpointQueryResult struct {
	// The request's journal.
	// Used to retrieve details of the JSONRPC request, e.g. the JSONRPC method.
	// Declared embedded to allow direct access by other members of the `judge` package.
	*requestJournal

	// Tracks the address of the endpoint for which the result is built.
	endpointAddr protocol.EndpointAddr

	// Tracks the payload received from the endpoint in response to the JSONRPC request.
	// Custom QoS service does NOT have access to this: it can only act on a parsed JSONRPC response.
	endpointPayload []byte

	// Captures the queried endpoint's error and the response to return to the client.
	// Can be set by either:
	// - JUDGE: e.g. if the endpoint's payload failed to parse as a JSONRPC response.
	// - Custom QoS: e.g. if the endpoint returned an unexpected block height.
	// Only set if the query result indicates an endpoint error.
	// It could also include sanctions:
	// e.g. for an invalid value returned for an EVM `eth_blockNumber` request, the custom service could set:
	// Error:
	// - Description: "invalid response to eth_blockNumber"
	// - RecommendedSanction: {Duration: 5 * time.Minute}
	EndpointError *EndpointError

	// Only set if the endpoint's returned payload could be parsed into a JSONRPC response.
	parsedJSONRPCResponse *jsonrpc.Response

	// The set of values/attributes extracted from the endpoint query and the endpoint's parsed JSONRPC response.
	// e.g. for a Solana `getEpochInfo` request, the custom service could derive two endpoint attributes as follows:
	// - "BlockHeight": 0x1234
	// - "Epoch": 5
	StrValues map[string]string
	IntValues map[string]int

	// The time at which the query result is expired.
	// Expired results will be ignored, including in:
	// - endpoint selection, e.g. sanctions.
	// - state update: e.g. archival state of the QoS service.
	ExpiryTime time.Time

	// TODO_FUTURE(@adshmh): add a JSONRPCErrorResponse to allow a result builder to supply its custom JSONRPC response.
}

func (eqr *EndpointQueryResult) GetEndpointAddr() protocol.EndpointAddr {
	return eqr.endpointAddr
}

func (eqr *EndpointQueryResult) IsJSONRPCError() bool {
	parsedJSONRPCResponse, err := eqr.getParsedJSONRPCResponse()
	if err != nil {
		return false
	}

	return parsedJSONRPCResponse.IsError()
}

func (eqr *EndpointQueryResult) GetResultAsInt() (int, error) {
	parsedJSONRPCResponse, err := eqr.getParsedJSONRPCResponse()
	if err != nil {
		return 0, err
	}

	return parsedJSONRPCResponse.GetResultAsInt()
}

func (eqr *EndpointQueryResult) GetResultAsStr() (string, error) {
	parsedJSONRPCResponse, err := eqr.getParsedJSONRPCResponse()
	if err != nil {
		return "", err
	}

	return parsedJSONRPCResponse.GetResultAsStr()
}

func (eqr *EndpointQueryResult) getParsedJSONRPCResponse() (*jsonrpc.Response, error) {
	parsedJSONRPCResponse := eqr.parsedJSONRPCResponse
	// Endpoint payload failed to parse as JSONRPC response.
	// This is not considered a JSONRPC error response.
	if parsedJSONRPCResponse == nil {
		return nil, fmt.Errorf("endpoint payload failed to parse as JSONRPC.")
	}

	return parsedJSONRPCResponse, nil
}

func (eqr *EndpointQueryResult) Success(
	resultBuilders ...ResultBuilder,
) *EndpointQueryResult {
	for _, builder := range resultBuilders {
		builder(eqr)
	}

	return eqr
}

// ErrorResult creates an error result with the given message and no sanction.
// Returns a self-reference for a fluent API.
func (eqr *EndpointQueryResult) Error(description string) *EndpointQueryResult {
	eqr.EndpointError = &EndpointError{
		ErrorKind: EndpointErrKindInvalidResult,
		// Description is set by the custom service implementation
		Description: description,
	}

	return eqr
}

// SanctionEndpoint creates an error result with a temporary sanction.
func (eqr *EndpointQueryResult) SanctionEndpoint(description, reason string, duration time.Duration) *EndpointQueryResult {
	eqr.EndpointError = &EndpointError{
		ErrorKind:   EndpointErrKindInvalidResult,
		Description: description,
		RecommendedSanction: &Sanction{
			Type:       SanctionTypeTemporary,
			Reason:     reason,
			ExpiryTime: time.Now().Add(duration),
		},
	}

	return eqr
}

// PermanentSanction creates an error result with a permanent sanction.
func (eqr *EndpointQueryResult) PermanentSanction(description, reason string) *EndpointQueryResult {
	eqr.EndpointError = &EndpointError{
		ErrorKind:   EndpointErrKindInvalidResult,
		Description: description,
		RecommendedSanction: &Sanction{
			Type:   SanctionTypePermanent,
			Reason: reason,
		},
	}

	return eqr
}

type ResultBuilder func(*EndpointQueryResult)

func (eqr *EndpointQueryResult) AddIntResult(key string, value int) ResultBuilder {
	return func(r *EndpointQueryResult) {
		if r.IntValues == nil {
			r.IntValues = make(map[string]int)
		}
		r.IntValues[key] = value
	}
}

func (eqr *EndpointQueryResult) AddStrResult(key, value string) ResultBuilder {
	return func(r *EndpointQueryResult) {
		if r.StrValues == nil {
			r.StrValues = make(map[string]string)
		}
		r.StrValues[key] = value
	}
}

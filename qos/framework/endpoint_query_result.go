package framework

import (
	"errors"
	"time"

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
	// The endpointQuery from which this result was built.
	// It can be used, e.g. to retrieve the JSONRPC request and its method.
	*endpointQuery

	// TODO_IN_THIS_PR: verify this is set by all result builders.

	// The JSONRPC response to be returned to the client.
	// MUST be set.
	clientResponse *jsonrpc.Response

	// The set of values/attributes extracted from the endpoint query and the endpoint's parsed JSONRPC response.
	// e.g. for a Solana `getEpochInfo` request, the custom service could derive two endpoint attributes as follows:
	// - "BlockHeight": 0x1234
	// - "Epoch": 5
	StringValues map[string]string
	IntValues    map[string]int

	// Captures the queried endpoint's error.
	// Only set if the query result indicates an endpoint error.
	// It could also include sanctions:
	// e.g. for an invalid value returned for an EVM `eth_blockNumber` request, the custom service could set:
	// Error:
	// - Description: "invalid response to eth_blockNumber"
	// - RecommendedSanction: {Duration: 5 * time.Minute}
	Error *EndpointError

	// The time at which the query result is expired.
	// Expired results will be ignored, including in:
	// - endpoint selection, e.g. sanctions.
	// - state update: e.g. archival state of the QoS service.
	ExpiryTime time.Time

	// TODO_FUTURE(@adshmh): add a JSONRPCErrorResponse to allow a result builder to supply its custom JSONRPC response.
}

// ===> These are moved from EndpointQueryResultContext --> it will have a public EndpointQueryResult field to allow direct access to the methods below.
func (eqr *EndpointQueryResult) IsJSONRPCError() bool {
	parsedJSONRPCResponse, err := eqr.getParsedJSONRPCResponse()
	if err != nil {
		return false
	}

	return parsedJSONRPCResponse.IsError()
}

func (eqr *EndpointQueryResult) GetResultAsInt() (int, error) {
	parsedJSONRPCResponse := eqr.getParsedJSONRPCResponse()
	if err != nil {
		return 0, err
	}

	return parsedJSONRPCResponse.GetResultAsInt()
}

func (ctx *EndpointQueryResultContext) GetResultAsStr() (string, error) {
	parsedJSONRPCResponse := eqr.getParsedJSONRPCResponse()
	if err != nil {
		return "", err
	}

	return parsedJSONRPCResponse.GetResultAsStr()
}

func (eqr *EndpointQueryResult) getParsedJSONRPCResponse() (*jsonrpc.Response, error) {
	parsedJSONRPCResponse := eqr.endpointQuery.parsedJSONRPCResponse
	// Endpoint payload failed to parse as JSONRPC response.
	// This is not considered a JSONRPC error response.
	if parsedJSONRPCResponse == nil {
		return nil, fmt.Errorf("endpoint payload failed to parse as JSONRPC.")
	}

	return parsedJSONRPCResponse, nil
}

func (eqr *EndpointQueryResult) AddIntValue(key string, value int) {
	eqr.endpointQueryResult.AddIntValue(key, value)
}

func (ctx *EndpointQueryResultContext) AddStrValue(key, value string) {
	ctx.endpointQueryResult.AddStrValue(key, value)
}


// ErrorResult creates an error result with the given message and no sanction.
func (ctx *EndpointQueryResultContext) ErrorResult(description string) *EndpointQueryResult {

	return &ResultData{
		Type: ctx.Method,
		Error: &ResultError{
			Description: description,
			kind:        EndpointDataErrorKindInvalidResult,
		},
	}
}

// SanctionEndpoint creates an error result with a temporary sanction.
func (ctx *EndpointQueryResultContext) SanctionEndpoint(description, reason string, duration time.Duration) *ResultData {
	return &ResultData{
		Type: ctx.Method,
		Error: &ResultError{
			Description: description,
			RecommendedSanction: &SanctionRecommendation{
				Sanction: Sanction{
					Type:        SanctionTypeTemporary,
					Reason:      reason,
					ExpiryTime:  time.Now().Add(duration),
					CreatedTime: time.Now(),
				},
				SourceDataType: ctx.Method,
				TriggerDetails: description,
			},
			kind: EndpointDataErrorKindInvalidResult,
		},
		CreatedTime: time.Now(),
	}
}

// PermanentSanction creates an error result with a permanent sanction.
func (ctx *EndpointQueryResultContext) PermanentSanction(description, reason string) *ResultData {
	return &ResultData{
		Type: ctx.Method,
		Error: &ResultError{
			Description: description,
			RecommendedSanction: &SanctionRecommendation{
				Sanction: Sanction{
					Type:        SanctionTypePermanent,
					Reason:      reason,
					CreatedTime: time.Now(),
				},
				SourceDataType: ctx.Method,
				TriggerDetails: description,
			},
			kind: EndpointDataErrorKindInvalidResult,
		},
		CreatedTime: time.Now(),
	}
}




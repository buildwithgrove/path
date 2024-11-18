package solana

import (
	"encoding/json"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// responseUnmarshallerBlockNumber deserializes the provided payload
// into a responseToBlockNumber struct, adding any encountered errors
// to the returned struct.
func responseUnmarshallerGetEpochInfo(jsonrpcReq jsonrpc.Request, jsonrpcResp jsonrpc.Response, logger polylog.Logger) (response, error) {
	if jsonrpcResp.Error.Code != 0 { // The endpoint returned an error: no need to do further processing of the response.
		// Note: this assumes the `getEpochInfo` request sent to the endpoint was valid.
		return responseToGetEpochInfo{
			Response: jsonrpcResp,
			Logger:   logger,
		}, nil
	}

	resultBz, err := jsonrpcResp.GetResultAsBytes()
	if err != nil {
		return responseToGetEpochInfo{
			Response: jsonrpcResp,
			Logger:   logger,
		}, err
	}

	var epochInfo epochInfo
	err = json.Unmarshal(resultBz, &epochInfo)

	return &responseToGetEpochInfo{
		Response:  jsonrpcResp,
		epochInfo: epochInfo,
	}, err
}

// epochInfo captures all the fields expected from a response to a `getEpochInfo` request.
type epochInfo struct {
	AbsoluteSlot     uint64 `json:"absoluteSlot"`
	BlockHeight      uint64 `json:"blockHeight"`
	Epoch            uint64 `json:"epoch"`
	SlotIndex        uint64 `json:"slotIndex"`
	SlotsInEpoch     uint64 `json:"slotsInEpoch"`
	TransactionCount uint64 `json:"transactionCount"`
}

// responseToGetEpochInfo captures the fields expected in a
// response to a `getEpochInfo` request.
type responseToGetEpochInfo struct {
	jsonrpc.Response
	Logger polylog.Logger

	// epochInfo stores the epochInfo struct that is parsed from the response to a `getEpochInfo` request.
	epochInfo epochInfo
}

// GetObservation returns a Solana Endpoint observation based on an endpoint's response to a `getEpochInfo` request.
// This method implements the response interface used by the requestContext struct. 
func (r responseToGetEpochInfo) GetObservation() observation.qos.SolanaEndpointDetails {
	return observation.qos.SolanaEndpointDetails{
		EpochInfo    : &observation.qos.SolanaEpochInfoResponse {
			BlockHeight: r.epochInfo.BlockHeight,
			Epoch: r.epochInfo.Epoch,
		},
	}
}

// TODO_UPNEXT(@adshmh): handle the following scenarios:
//  1. An endpoint returned a malformed, i.e. Not in JSONRPC format, response.
//     The user-facing response should include the request's ID.
//  2. An endpoint returns a JSONRPC response indicating a user error:
//     This should be returned to the user as-is.
//  3. An endpoint returns a valid JSONRPC response to a valid user request:
//     This should be returned to the user as-is.
func (r responseToGetEpochInfo) GetResponsePayload() []byte {
	bz, err := json.Marshal(r.Response)
	if err != nil {
		// This should never happen: log an entry but return the response anyway.
		r.Logger.Warn().Err(err).Msg("responseToGetEpochInfo: Marshalling JSONRPC response failed.")
	}
	return bz
}

// epochInfoResponseObservation provides the functionality defined by the response interface, specific to a response matching
// a `getEpochInfo` request.
var _ observation = epochInfoResponseObservation{}

// epochInfoResponseObservation holds the epochInfo struct built from a response, and applies it to the (supplied) corresponding endpoint's struct.
type epochInfoResponseObservation struct {
	epochInfo epochInfo
}

func (e epochInfoResponseObservation) Apply(ep *endpoint) {
	ep.GetEpochInfoResult = &e.epochInfo
}

package jsonrpc

import (
	"encoding/json"
)

type BatchRequest struct {
	Requests []Request
}

// TODO_UPNEXT(@adshmh): Validate ID values, e.g. for duplicate values, when unmarshaling.
//
// GetRequestPayloads returns the slice of serialized forms of JSONRPC requests.
func (br *BatchRequest) GetRequestsPayloads() [][]byte {
	requestPayloads := make([][]byte, len(br.Requests))
	for i, req := range br.Requests {
		// TODO_TECHDEBT(@adshmh): Log an entry if there is an error marshaling.
		// A marshaling error here should never happen here.
		payload, _ := json.Marshal(req)
		requestPayloads[i] = payload
	}

	return requestPayloads
}

// Custom unmarshaller to support requests of the format `[{"jsonrpc":"2.0","id":1},{"jsonrpc":"2.0","id":2}]`
func (br *BatchRequest) UnmarshalJSON(data []byte) error {
	var requests []Request
	if err := json.Unmarshal(data, &requests); err != nil {
		return err
	}

	br.Requests = requests
	return nil
}

// TODO_UPNEXT(@adshmh): Validate responses in the batch
//
// BuildResponseBytes constructs a Batch JSONRPC response from the slice of response payloads.
func (br *BatchRequest) BuildResponseBytes(jsonrpcResponses []Response) []byte {
	// TODO_TECHDEBT(@adshmh): Refactor so marshaling a Response never fails.
	responseBz, _ := json.Marshal(jsonrpcResponses)

	return responseBz
}

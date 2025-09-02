package jsonrpc

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pokt-network/poktroll/pkg/polylog"
)

// Batch validation errors - standard Go error variables
var (
	ErrBatchResponseLengthMismatch = errors.New("batch response length mismatch")
	ErrBatchResponseMissingIDs     = errors.New("batch response missing required IDs")
	ErrBatchResponseEmpty          = errors.New("empty batch response not allowed per JSON-RPC specification")
	ErrBatchResponseMarshalFailure = errors.New("failed to marshal batch response")
)

// ValidateAndBuildBatchResponse validates and constructs a batch response according to JSON-RPC 2.0 specification.
// It performs comprehensive validation including:
//   - Empty batch handling per JSON-RPC spec (returns empty payload)
//   - Response length matches request length
//   - All request IDs are present in responses
//   - Proper JSON array construction
//
// Returns the marshaled JSON byte array for the response payload.
// For empty batches, returns empty byte slice per JSON-RPC spec requirement to return "nothing at all".
func ValidateAndBuildBatchResponse(
	logger polylog.Logger,
	responses []json.RawMessage,
	jsonrpcReqs map[string]Request,
) ([]byte, error) {

	// Handle empty batch according to JSON-RPC spec first
	// "If there are no Response objects contained within the Response array
	// as it is to be sent to the client, the server MUST NOT return an empty Array and should return nothing at all."
	if len(responses) == 0 {
		logger.Debug().Msg("Batch request resulted in no response objects - returning empty response per JSON-RPC spec")
		return []byte{}, nil // Return empty payload representing "nothing at all"
	}

	// Validate response length matches request length
	if err := validateResponseLength(responses, jsonrpcReqs); err != nil {
		return nil, err
	}

	// Validate all request IDs are present in responses
	if err := validateResponseIDs(responses, jsonrpcReqs); err != nil {
		return nil, err
	}

	// Marshal responses into JSON array
	return marshalBatchResponse(responses)
}

// validateResponseLength ensures response count matches request count
func validateResponseLength(responses []json.RawMessage, jsonrpcReqs map[string]Request) error {
	if len(responses) != len(jsonrpcReqs) {
		return fmt.Errorf("%w: expected %d responses, got %d",
			ErrBatchResponseLengthMismatch, len(jsonrpcReqs), len(responses))
	}
	return nil
}

// validateResponseIDs ensures all request IDs are present in the responses
func validateResponseIDs(responses []json.RawMessage, jsonrpcReqs map[string]Request) error {
	// Parse response IDs from the raw JSON messages
	responseIDs := make(map[string]bool)
	for _, respMsg := range responses {
		var resp Response
		if err := json.Unmarshal(respMsg, &resp); err != nil {
			// Skip invalid responses - they'll be handled elsewhere
			continue
		}
		responseIDs[resp.ID.String()] = true
	}

	// Check for missing request IDs in responses
	var missingIDs []string
	for reqID := range jsonrpcReqs {
		if !responseIDs[reqID] {
			missingIDs = append(missingIDs, reqID)
		}
	}

	if len(missingIDs) > 0 {
		return fmt.Errorf("%w: %v", ErrBatchResponseMissingIDs, missingIDs)
	}

	return nil
}

// marshalBatchResponse constructs the final JSON array from individual responses
func marshalBatchResponse(responses []json.RawMessage) ([]byte, error) {
	batchResponse, err := json.Marshal(responses)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBatchResponseMarshalFailure, err)
	}
	return batchResponse, nil
}

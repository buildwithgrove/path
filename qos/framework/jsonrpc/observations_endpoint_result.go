package framework

import (
	"github.com/buildwithgrove/path/observation/qos/jsonrpc"
	jsonrpcobservations "github.com/buildwithgrove/path/observation/qos/framework"
	"github.com/buildwithgrove/path/qos/jsonrpc"
)

// buildObservations converts an EndpointQueryResult to jsonrpc.Observations
// for reporting metrics and analysis.
func (eqr *EndpointQueryResult) buildObservations() *observations.EndpointQueryResult {
	// Create the endpoint attributes map for the protobuf message
	protoAttributes := make(map[string]*observations.EndpointAttribute)

	// Convert each EndpointAttribute to its protobuf representation
	for key, attr := range eqr.EndpointAttributes {
		protoAttr := &observations.EndpointAttribute{}
		
		// Convert the value based on its type (string or int)
		if strVal, ok := attr.GetStringValue(); ok {
			protoAttr.Value = &observations.EndpointAttribute_StringValue{
				StringValue: strVal,
			}
		} else if intVal, ok := attr.GetIntValue(); ok {
			protoAttr.Value = &observations.EndpointAttribute_IntValue{
				IntValue: int32(intVal),
			}
		}
		
		// Add error information if present
		if attr.err != nil {
			protoAttr.Error = &observations.EndpointAttributeError{
				ErrorKind:   observations.EndpointErrorKind(attr.err.kind),
				Description: attr.err.Description,
			}
			
			// Add sanction information if present
			if attr.err.RecommendedSanction != nil {
				protoAttr.Error.RecommendedSanction = &observations.Sanction{
					Type:   observations.SanctionType(attr.err.RecommendedSanction.Type),
					Reason: attr.err.RecommendedSanction.Reason,
				}
				
				// Convert expiry time if it's not zero
				if !attr.err.RecommendedSanction.ExpiryTime.IsZero() {
					ts, _ := ptypes.TimestampProto(attr.err.RecommendedSanction.ExpiryTime)
					protoAttr.Error.RecommendedSanction.ExpiryTimestamp = ts
				}
			}
		}
		
		protoAttributes[key] = protoAttr
	}
	
	// Create and return the EndpointQueryResult proto
	return &observations.EndpointQueryResult{
		EndpointAttributes: protoAttributes,
	}
}

// extractEndpointAttributes converts protobuf Observations to []*EndpointQueryResult
// This allows the framework to recreate EndpointQueryResults from serialized observations
func extractEndpointQueryResults(observations *observations.Observations) []*EndpointQueryResult {
	results := make([]*EndpointQueryResult, 0, len(observations.EndpointObservations))
	
	// Extract data from each endpoint observation
	for _, observation := range observations.EndpointObservations {
		// Create a new EndpointQueryResult
		result := &EndpointQueryResult{
			EndpointAttributes: make(map[string]jsonrpc.EndpointAttribute),
		}
		
		// Convert protobuf attributes to EndpointAttribute objects
		for key, protoAttr := range observation.Result.EndpointAttributes {
			attr := jsonrpc.EndpointAttribute{}
			
			// Convert value based on type
			switch v := protoAttr.Value.(type) {
			case *observations.EndpointAttribute_StringValue:
				strVal := v.StringValue
				attr.stringValue = &strVal
			case *observations.EndpointAttribute_IntValue:
				intVal := int(v.IntValue)
				attr.intValue = &intVal
			}
			
			// Convert error information if present
			if protoAttr.Error != nil {
				attr.err = &jsonrpc.EndpointAttributeError{
					Description: protoAttr.Error.Description,
					kind:        jsonrpc.endpointErrorKind(protoAttr.Error.ErrorKind),
				}
				
				// Convert sanction information if present
				if protoAttr.Error.RecommendedSanction != nil {
					attr.err.RecommendedSanction = &jsonrpc.Sanction{
						Type:   jsonrpc.SanctionType(protoAttr.Error.RecommendedSanction.Type),
						Reason: protoAttr.Error.RecommendedSanction.Reason,
					}
					
					// Convert expiry timestamp if present
					if protoAttr.Error.RecommendedSanction.ExpiryTimestamp != nil {
						t, _ := ptypes.Timestamp(protoAttr.Error.RecommendedSanction.ExpiryTimestamp)
						attr.err.RecommendedSanction.ExpiryTime = t
					}
				}
			}
			
			// Add attribute to the result
			result.EndpointAttributes[key] = attr
		}
		
		results = append(results, result)
	}
	
	return results
}

package devtools

import (
	"time"

	"github.com/buildwithgrove/path/protocol"
)

// DisqualifiedEndpointResponse contains useful information about currently disqualified endpoints.
// It reports data from both the:
//   - Protocol-level disqualified endpoints
//   - QoS-level disqualified endpoints
//
// It also reports the total number of service endpoints, the number of qualified service endpoints, and the number of disqualified service endpoints.
type DisqualifiedEndpointResponse struct {
	ProtocolLevelDisqualifiedEndpoints ProtocolLevelDataResponse `json:"protocol_level_disqualified_endpoints"`
	QoSLevelDisqualifiedEndpoints      QoSLevelDataResponse      `json:"qos_level_disqualified_endpoints"`
	TotalServiceEndpointsCount         int                       `json:"total_service_endpoints_count"`
	QualifiedServiceEndpointsCount     int                       `json:"qualified_service_endpoints_count"`
	DisqualifiedServiceEndpointsCount  int                       `json:"disqualified_service_endpoints_count"`
}

// GetDisqualifiedEndpointsCount sums:
//   - Protocol-level permanently sanctioned endpoints
//   - Protocol-level session sanctioned endpoints
//   - QoS-level disqualified endpoints
func (r *DisqualifiedEndpointResponse) GetDisqualifiedEndpointsCount() int {
	return len(r.ProtocolLevelDisqualifiedEndpoints.PermanentlySanctionedEndpoints) +
		len(r.ProtocolLevelDisqualifiedEndpoints.SessionSanctionedEndpoints) +
		len(r.QoSLevelDisqualifiedEndpoints.DisqualifiedEndpoints)
}

// GetValidServiceEndpointsCount subtracts the number of disqualified endpoints from the total number of service endpoints.
func (r *DisqualifiedEndpointResponse) GetValidServiceEndpointsCount() int {
	return r.TotalServiceEndpointsCount - r.GetDisqualifiedEndpointsCount()
}

type (
	// ProtocolLevelDataResponse contains data about sanctioned endpoints at the protocol level.
	// It reports the number of permanently sanctioned endpoints, the number of session sanctioned endpoints, and the total number of sanctioned endpoints.
	ProtocolLevelDataResponse struct {
		PermanentlySanctionedEndpoints    map[protocol.EndpointAddr]SanctionedEndpoint `json:"permanently_sanctioned_endpoints"`
		SessionSanctionedEndpoints        map[protocol.EndpointAddr]SanctionedEndpoint `json:"session_sanctioned_endpoints"`
		PermamentSanctionedEndpointsCount int                                          `json:"permanent_sanctioned_endpoints_count"`
		SessionSanctionedEndpointsCount   int                                          `json:"session_sanctioned_endpoints_count"`
		TotalSanctionedEndpointsCount     int                                          `json:"total_sanctioned_endpoints_count"`
	}

	// QoSLevelDataResponse contains data about disqualified endpoints at the QoS level.
	// It reports the number of disqualified endpoints, the number of empty response endpoints, the number of chain ID check errors, the number of archival check errors, and the number of block number check errors.
	QoSLevelDataResponse struct {
		DisqualifiedEndpoints       map[protocol.EndpointAddr]QoSDisqualifiedEndpoint `json:"disqualified_endpoints"`
		EmptyResponseCount          int                                               `json:"empty_response_count"`
		ChainIDCheckErrorsCount     int                                               `json:"chain_id_check_errors_count"`
		ArchivalCheckErrorsCount    int                                               `json:"archival_check_errors_count"`
		BlockNumberCheckErrorsCount int                                               `json:"block_number_check_errors_count"`
	}

	// SanctionedEndpoint represents an endpoint sanctioned at the protocol level.
	SanctionedEndpoint struct {
		SupplierAddress string `json:"supplier_address"`
		EndpointURL     string `json:"endpoint_url"`
		// SessionID is only set for session-based sanctions.
		SessionID     string             `json:"session_id,omitempty"`
		ServiceID     protocol.ServiceID `json:"service_id"`
		Reason        string             `json:"reason"`
		SanctionType  string             `json:"sanction_type"`
		ErrorType     string             `json:"error_type"`
		SessionHeight int64              `json:"session_height"`
		CreatedAt     time.Time          `json:"created_at"`
	}

	// QoSDisqualifiedEndpoint represents an endpoint disqualified at the QoS level.
	QoSDisqualifiedEndpoint struct {
		SupplierAddress string             `json:"supplier_address"`
		EndpointURL     string             `json:"endpoint_url"`
		Reason          string             `json:"reason"`
		ServiceID       protocol.ServiceID `json:"service_id"`
	}
)

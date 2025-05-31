package devtools

import (
	"time"

	"github.com/buildwithgrove/path/protocol"
)

type DisqualifiedEndpointResponse struct {
	ProtocolLevelDataResponse    ProtocolLevelDataResponse `json:"protocol_level_data_response"`
	QoSLevelDataResponse         QoSLevelDataResponse      `json:"qos_level_data_response"`
	TotalServiceEndpointsCount   int                       `json:"total_service_endpoints_count"`
	ValidServiceEndpointsCount   int                       `json:"valid_service_endpoints_count"`
	InvalidServiceEndpointsCount int                       `json:"invalid_service_endpoints_count"`
}

func (r *DisqualifiedEndpointResponse) GetDisqualifiedEndpointsCount() int {
	return len(r.ProtocolLevelDataResponse.PermanentlySanctionedEndpoints) +
		len(r.ProtocolLevelDataResponse.SessionSanctionedEndpoints) +
		len(r.QoSLevelDataResponse.DisqualifiedEndpoints)
}

func (r *DisqualifiedEndpointResponse) GetValidServiceEndpointsCount() int {
	return r.TotalServiceEndpointsCount - r.GetDisqualifiedEndpointsCount()
}

// ProtocolLevelDataResponse is the response from the GetSanctionedEndpoints function.
type (
	ProtocolLevelDataResponse struct {
		PermanentlySanctionedEndpoints    map[string]SanctionedEndpoint `json:"permanently_sanctioned_endpoints"`
		SessionSanctionedEndpoints        map[string]SanctionedEndpoint `json:"session_sanctioned_endpoints"`
		PermamentSanctionedEndpointsCount int                           `json:"permanent_sanctioned_endpoints_count"`
		SessionSanctionedEndpointsCount   int                           `json:"session_sanctioned_endpoints_count"`
		TotalSanctionedEndpointsCount     int                           `json:"total_sanctioned_endpoints_count"`
	}

	QoSLevelDataResponse struct {
		DisqualifiedEndpoints       map[protocol.EndpointAddr]DisqualifiedEndpoint `json:"disqualified_endpoints"`
		EmptyResponseCount          int                                            `json:"empty_response_count"`
		ChainIDCheckErrorsCount     int                                            `json:"chain_id_check_errors_count"`
		ArchivalCheckErrorsCount    int                                            `json:"archival_check_errors_count"`
		BlockNumberCheckErrorsCount int                                            `json:"block_number_check_errors_count"`
	}

	SanctionedEndpoint struct {
		EndpointAddr  protocol.EndpointAddr `json:"endpoint_addr"`
		Reason        string                `json:"reason"`
		ServiceID     protocol.ServiceID    `json:"service_id"`
		SanctionType  string                `json:"sanction_type"`
		ErrorType     string                `json:"error_type"`
		SessionHeight int64                 `json:"session_height"`
		CreatedAt     time.Time             `json:"created_at"`
	}

	// DisqualifiedEndpoint is the details of a sanctioned endpoint.
	DisqualifiedEndpoint struct {
		EndpointAddr protocol.EndpointAddr `json:"endpoint_addr"`
		Reason       string                `json:"reason"`
		ServiceID    protocol.ServiceID    `json:"service_id"`
	}
)

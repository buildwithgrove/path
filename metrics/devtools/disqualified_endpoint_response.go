package devtools

import (
	"time"

	"github.com/buildwithgrove/path/protocol"
)

type DisqualifiedEndpointResponse struct {
	ProtocolLevelDataResponse ProtocolLevelDataResponse `json:"protocol_level_data_response"`
	QoSLevelDataResponse      QoSLevelDataResponse      `json:"qos_level_data_response"`
	AvailableEndpointsCount   int                       `json:"available_endpoints_count"`
	ValidEndpointsCount       int                       `json:"valid_endpoints_count"`
	InvalidEndpointsCount     int                       `json:"invalid_endpoints_count"`
}

func (r *DisqualifiedEndpointResponse) GetDisqualifiedEndpointsCount() int {
	return len(r.ProtocolLevelDataResponse.PermanentlySanctionedEndpoints) +
		len(r.ProtocolLevelDataResponse.SessionSanctionedEndpoints) +
		len(r.QoSLevelDataResponse.DisqualifiedEndpoints)
}

// ProtocolLevelDataResponse is the response from the GetSanctionedEndpoints function.
// It exists to allow quick information about currently sanctioned endpoints and
// will eventually be removed in favour of a metrics-based approach.
type ProtocolLevelDataResponse struct {
	PermanentlySanctionedEndpoints    map[protocol.EndpointAddr]DisqualifiedEndpoint `json:"permanently_sanctioned_endpoints"`
	SessionSanctionedEndpoints        map[protocol.EndpointAddr]DisqualifiedEndpoint `json:"session_sanctioned_endpoints"`
	SanctionedEndpointsCount          int                                            `json:"sanctioned_endpoints_count"`
	PermamentSanctionedEndpointsCount int                                            `json:"permanent_sanctioned_endpoints_count"`
	SessionSanctionedEndpointsCount   int                                            `json:"session_sanctioned_endpoints_count"`
}

type QoSLevelDataResponse struct {
	DisqualifiedEndpoints       map[protocol.EndpointAddr]DisqualifiedEndpoint `json:"disqualified_endpoints"`
	EmptyResponseCount          int                                            `json:"empty_response_count"`
	ChainIDCheckErrorsCount     int                                            `json:"chain_id_check_errors_count"`
	ArchivalCheckErrorsCount    int                                            `json:"archival_check_errors_count"`
	BlockNumberCheckErrorsCount int                                            `json:"block_number_check_errors_count"`
}

// DisqualifiedEndpoint is the details of a sanctioned endpoint.
// It exists to allow quick information about currently sanctioned endpoints and
// will eventually be removed in favour of a metrics-based approach.
type DisqualifiedEndpoint struct {
	EndpointAddr protocol.EndpointAddr `json:"endpoint_addr"`
	Reason       string                `json:"reason"`
	ServiceID    protocol.ServiceID    `json:"service_id"`

	// Sanctions only - TODO_IN_THIS_PR: move to separate struct?
	SanctionType  string    `json:"sanction_type"`
	ErrorType     string    `json:"error_type"`
	SessionHeight int64     `json:"session_height"`
	CreatedAt     time.Time `json:"created_at"`
}

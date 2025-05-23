package devtools

import (
	"time"

	"github.com/buildwithgrove/path/protocol"
)

// SanctionDetailsResponse is the response from the GetSanctionedEndpoints function.
// It exists to allow quick information about currently sanctioned endpoints and
// will eventually be removed in favour of a metrics-based approach.
type SanctionDetailsResponse struct {
	PermamentSanctionDetails          map[protocol.ServiceID][]SanctionDetails `json:"permanent_sanction_details"`
	SessionSanctionDetails            map[protocol.ServiceID][]SanctionDetails `json:"session_sanction_details"`
	TotalEndpointsCount               int                                      `json:"total_endpoints_count"`
	ValidEndpointsCount               int                                      `json:"valid_endpoints_count"`
	SanctionedEndpointsCount          int                                      `json:"sanctioned_endpoints_count"`
	PermamentSanctionedEndpointsCount int                                      `json:"permanent_sanctioned_endpoints_count"`
	SessionSanctionedEndpointsCount   int                                      `json:"session_sanctioned_endpoints_count"`
}

// SanctionDetails is the details of a sanctioned endpoint.
// It exists to allow quick information about currently sanctioned endpoints and
// will eventually be removed in favour of a metrics-based approach.
type SanctionDetails struct {
	EndpointAddr  protocol.EndpointAddr `json:"endpoint_addr"`
	Reason        string                `json:"reason"`
	SanctionType  string                `json:"sanction_type"`
	ErrorType     string                `json:"error_type"`
	ServiceID     protocol.ServiceID    `json:"service_id"`
	SessionHeight int                   `json:"session_height"`
	CreatedAt     time.Time             `json:"created_at"`
}

package request

import "github.com/pokt-foundation/portal-middleware/relayer"

// SupportedServicesToQoSServiceName is a map of service IDs to the QoS Service that will be used
// to perform request parsing, response building, and endpoint selection for the given service ID.
var supportedServicesToQoSServiceName = map[relayer.ServiceID]ServiceName{

	"gatewaye2e": ServiceNameEVM, // "gatewaye2e" service used in E2E tests

	"0021": ServiceNameEVM,
	"0040": ServiceNameEVM,

	"0006": ServiceNameSolana,
	"C006": ServiceNameSolana,

	"0001": ServiceNamePOKT,

	// TODO_IMPROVE: add all supported service IDs and their corresponding QoS service types
}

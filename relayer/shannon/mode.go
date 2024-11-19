package shannon

import (
	"net/http"

	apptypes "github.com/pokt-network/poktroll/x/application/types"
)

// OperationMode defines the behavior of a specific mode of operation of PATH.
// As of now, it is expected to provide a single method to build an instance of the underlying mode of operation based on an HTTP request.
//
// TODO_DOCUMENT(@adshmh): Convert the following notion doc into a proper README.
// See the following link for more details on PATH's different modes of operation.
// https://www.notion.so/buildwithgrove/Different-Modes-of-Operation-PATH-LocalNet-Discussions-122a36edfff6805e9090c9a14f72f3b5
type OperationMode interface {
	BuildInstanceFromHTTPRequest(*http.Request) (OperationModeInstance, error)
}

// OperationModeInstance defines the behavior expected from an instance of any of the PATH's modes of operation.
// Any such instance is expected to provide the following:
//  1. App(s) on behalf of which a relay request can be sent.
//  2. A signer for signing a relay request. As of now, the operation mode instances return a signer that signs relay requests using either:
//     a. The private key of the gateway to which an app delegates (Centralized mode of operation)
//     b. The private key of the app selected by the user for the specific relay request (Trusted mode of operation)
type OperationModeInstance interface {
	GetAppFilterFn() IsAppPermittedFn
	GetRelayRequestSigner() RelayRequestSigner
}

// TODO_TECHDEBT: once Shannon supports querying the applications based on one more criteria, this function's name and signature should be updated to
// build and return the query criteria.
// e.g. a Centralized Operation Mode instance will return the required criteria to only select apps which delegate to its configured gateway address.
//
// IsAppPermittedFn is a function, supplied by an instance of an operation mode, that determines whether an application is permitted to be used for sending relays
// in the context of the specific operation mode instance.
type IsAppPermittedFn func(*apptypes.Application) bool

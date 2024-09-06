package morse

import (
	"github.com/pokt-foundation/pocket-go/provider"

	"github.com/buildwithgrove/path/relayer"
)

// The relayer package's App interface is fulfilled by the app struct below.
// app contains the additional fields publicKey and aat, built from Morse onchain data,
// which are only used by this package to send relays to Morse endpoints.
var _ relayer.App = app{}

// app contains the fields necessary to identify a Morse app, and use it in sending relays.
type app struct {
	address string
	// We use the application's address to identify it, but the publicKey is needed to get a session for a Morse application.
	publicKey string
	// aat is needed for signing relays sent on behalf of the Morse application.
	aat provider.PocketAAT
}

func (a app) Addr() relayer.AppAddr {
	return relayer.AppAddr(a.address)
}

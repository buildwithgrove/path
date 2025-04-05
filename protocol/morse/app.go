package morse

import (
	"github.com/pokt-foundation/pocket-go/provider"
)

// app contains the fields necessary to identify a Morse app, and use it in sending relays.
type app struct {
	address string
	// We use the application's address to identify it, but the publicKey is needed to get a session for a Morse application.
	publicKey string
	// aat is needed for signing relays sent on behalf of the Morse application.
	aat provider.PocketAAT
}

func (a app) IsEmpty() bool {
	return a.address == "" || a.publicKey == "" || a.aat.Signature == ""
}

func (a app) Addr() string {
	return a.address
}

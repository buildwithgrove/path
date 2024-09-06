package shannon

import (
	"github.com/pokt-foundation/portal-middleware/relayer"
)

// relayer package's App interface is fulfilled by the app struct below.
// app uses the onchain address of a Shannon application as its unique identifier.
var _ relayer.App = app{}

// app is used to build a relayer package App from a Shannon Application.
type app struct {
	address string
}

func (a app) Addr() relayer.AppAddr {
	return relayer.AppAddr(a.address)
}

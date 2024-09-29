package jsonrpc

type Method string
type Version string

const Version2 = Version("2.0")

type Request struct {
	ID      `jsonrpc:"id"`
	JSONRPC Version `jsonrpc:"jsonrpc"`
	Method  `jsonrpc:"method"`
	// TODO_TECHDEBT: support other forms of params field, based on the JSONRPC spec.
	// See the link below for more details:
	// https://www.jsonrpc.org/specification
	//
	// For an example of a JSONRPC request where the params field
	// is not a slice of strings, see the link below:
	// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_signtransaction
	Params []string `jsonrpc:"params"`
}

const (
	// TODO_IMPROVE: return the same request ID as the request that caused the error
	evmErrorTemplate = `{"jsonrpc":"2.0","id":"0","error":{"code":-32603,"message":"%s"}}`
)

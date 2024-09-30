package jsonrpc

type Method string
type Version string

const Version2 = Version("2.0")

type Request struct {
	ID      ID      `json:"id,omitempty"`
	JSONRPC Version `json:"jsonrpc"`
	Method  Method  `json:"method"`
	// TODO_TECHDEBT: support other forms of params field, based on the JSONRPC spec.
	// See the link below for more details:
	// https://www.jsonrpc.org/specification
	//
	// For an example of a JSONRPC request where the params field
	// is not a slice of strings, see the link below:
	// https://ethereum.org/en/developers/docs/apis/json-rpc/#eth_signtransaction
	Params []string `json:"params,omitempty"`
}

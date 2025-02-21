package jsonrpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO_MVP(@adshmh): add a test case for batch JSONRPC requests
func TestUnmarshalParams(t *testing.T) {
	testCases := []struct {
		name       string
		rawPayload []byte
	}{
		{
			name:       "params as array of single object",
			rawPayload: []byte(`{"jsonrpc":"2.0","id":8522963871549545,"method":"eth_getLogs","params":[{"address":["0x1234","0xabcd","0xffff"],"fromBlock":"0x1234567","toBlock":"0xfffffff","topics":[]}]}`),
		},
		{
			name:       "params as array of multiple objects",
			rawPayload: []byte(`{"jsonrpc":"2.0","id":1949014,"method":"eth_call","params":[{"data":"0x12345678","from":"0x0000000000000000000000000000000000000000","to":"0x12345678abcdef12345678abcdef12345678abcd"},"latest"]}`),
		},
		{
			name:       "params as mix of string and boolean",
			rawPayload: []byte(`{"id":1,"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false]}`),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := json.Unmarshal(testCase.rawPayload, &Request{})
			require.NoError(t, err)
		})
	}
}

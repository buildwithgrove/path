package evm

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	//	"github.com/buildwithgrove/path/gateway"
)

func TestParseHTTPRequest(t *testing.T) {
	testCases := []struct {
		desc        string
		requestBody string
	}{
		{
			desc:        "eth_chainId",
			requestBody: `{"id": 1001, "jsonrpc": "2.0", "method": ""}`,
			// requestBody: `{"id": 1001, "method": "", "params": []}`
		},
	}

	for _, tc := range testCases {
		qos := &QoS{endpointStore: &EndpointStore{}}
		httpReq := httptest.NewRequestWithContext(context.TODO(), "POST", "/", strings.NewReader(tc.requestBody))

		qosContext, isValid := qos.ParseHTTPRequest(context.TODO(), httpReq)
		t.Fatalf("isValid: %t, %v\n", isValid, string(qosContext.GetServicePayload().Data))
	}

}

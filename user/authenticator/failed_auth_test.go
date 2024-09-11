package authenticator

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_failedAuth(t *testing.T) {
	tests := []struct {
		name            string
		resp            *failedAuth
		expectedBody    []byte
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name:            "should return correct values",
			resp:            &failedAuth{body: "there was a button. I pushed it."},
			expectedBody:    []byte("there was a button. I pushed it."),
			expectedStatus:  http.StatusUnauthorized,
			expectedHeaders: map[string]string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			c.Equal(test.expectedBody, test.resp.GetPayload())
			c.Equal(test.expectedStatus, test.resp.GetHTTPStatusCode())
			c.Equal(test.expectedHeaders, test.resp.GetHTTPHeaders())
		})
	}
}

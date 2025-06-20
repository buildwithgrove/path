package router

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"

	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/health"
)

func newTestRouter(t *testing.T) (*router, *Mockgateway, *httptest.Server) {
	ctrl := gomock.NewController(t)
	mockGateway := NewMockgateway(ctrl)
	mockDisqualifiedEndpointsReporter := NewMockdisqualifiedEndpointsReporter(ctrl)

	r := NewRouter(
		polyzero.NewLogger(),
		mockGateway,
		mockDisqualifiedEndpointsReporter,
		&health.Checker{},
		config.RouterConfig{},
	)
	ts := httptest.NewServer(r.mux)
	t.Cleanup(ts.Close)

	return r, mockGateway, ts
}

func Test_handleHealthz(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "should return 200 with status ok",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"ready","imageTag":"development"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			_, _, ts := newTestRouter(t)

			// Create request
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/healthz", ts.URL), nil)
			c.NoError(err)

			// Perform request
			client := &http.Client{}
			resp, err := client.Do(req)
			c.NoError(err)
			defer resp.Body.Close()

			// Test assertions
			c.Equal(test.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			c.NoError(err)
			c.JSONEq(test.expectedBody, string(body))
		})
	}
}

func Test_handleHTTPServiceRequest(t *testing.T) {
	tests := []struct {
		name           string
		payload        string
		expectedBytes  []byte
		expectedStatus int
		expectedError  error
		path           string
		expectedPath   string
	}{
		{
			name:           "should perform a service request successfully",
			payload:        `{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber"}`,
			expectedBytes:  []byte(`{"jsonrpc": "2.0", "id": 1, "result": "0x10d4f"}`),
			expectedStatus: http.StatusOK,
			path:           "/v1",
			expectedPath:   "",
		},
		{
			name:           "should fail if service request handler returns an error",
			payload:        `{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber"}`,
			expectedBytes:  []byte("failed to send service request: some error\n"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  errors.New("some error"),
			path:           "/v1/",
			expectedPath:   "/",
		},
		{
			name:           "should handle /v1/whatever",
			payload:        `{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber"}`,
			expectedBytes:  []byte(`{"jsonrpc": "2.0", "id": 1, "result": "0x10d4f"}`),
			expectedStatus: http.StatusOK,
			path:           "/v1/whatever",
			expectedPath:   "/whatever",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)

			_, mockGateway, ts := newTestRouter(t)

			mockGateway.EXPECT().HandleServiceRequest(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, req *http.Request, w http.ResponseWriter) error {
					if req.URL.Path != test.expectedPath {
						t.Errorf("expected path %s, got %s", test.expectedPath, req.URL.Path)
					}
					if test.expectedStatus == http.StatusOK {
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write(test.expectedBytes)
						return nil
					} else {
						http.Error(w, "failed to send service request: some error", http.StatusInternalServerError)
					}
					return test.expectedError
				},
			)

			req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", ts.URL, test.path), strings.NewReader(test.payload))
			c.NoError(err)

			client := &http.Client{}
			resp, err := client.Do(req)
			c.NoError(err)
			defer resp.Body.Close()

			c.Equal(test.expectedStatus, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			c.NoError(err)
			c.Equal(test.expectedBytes, body)
		})
	}
}

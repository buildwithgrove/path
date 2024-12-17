package auth

import (
	"testing"

	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
)

func Test_URLPathExtractor(t *testing.T) {
	tests := []struct {
		name    string
		request *envoy_auth.AttributeContext_HttpRequest
		want    string
		wantErr bool
	}{
		{
			name: "should extract endpoint ID from valid path",
			request: &envoy_auth.AttributeContext_HttpRequest{
				Path: "/v1/1a2b3c4d",
			},
			want:    "1a2b3c4d",
			wantErr: false,
		},
		{
			name: "should return error for path without endpoint ID",
			request: &envoy_auth.AttributeContext_HttpRequest{
				Path: "/v1/",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "should return error for invalid path",
			request: &envoy_auth.AttributeContext_HttpRequest{
				Path: "/invalid/1a2b3c4d",
			},
			want:    "",
			wantErr: true,
		},
	}

	extractor := &URLPathExtractor{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := extractor.extractGatewayEndpointID(test.request)
			if (err != nil) != test.wantErr {
				t.Errorf("extractGatewayEndpointID() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if got != test.want {
				t.Errorf("extractGatewayEndpointID() = %v, want %v", got, test.want)
			}
		})
	}
}

func Test_HeaderExtractor(t *testing.T) {
	tests := []struct {
		name    string
		request *envoy_auth.AttributeContext_HttpRequest
		want    string
		wantErr bool
	}{
		{
			name: "should extract endpoint ID from header",
			request: &envoy_auth.AttributeContext_HttpRequest{
				Headers: map[string]string{
					"x-endpoint-id": "1a2b3c4d",
				},
			},
			want:    "1a2b3c4d",
			wantErr: false,
		},
		{
			name: "should return error if header is missing",
			request: &envoy_auth.AttributeContext_HttpRequest{
				Headers: map[string]string{},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "should return error if header is empty",
			request: &envoy_auth.AttributeContext_HttpRequest{
				Headers: map[string]string{
					"x-endpoint-id": "",
				},
			},
			want:    "",
			wantErr: true,
		},
	}

	extractor := &HeaderExtractor{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := extractor.extractGatewayEndpointID(test.request)
			if (err != nil) != test.wantErr {
				t.Errorf("extractGatewayEndpointID() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if got != test.want {
				t.Errorf("extractGatewayEndpointID() = %v, want %v", got, test.want)
			}
		})
	}
}

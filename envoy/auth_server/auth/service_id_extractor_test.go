package auth

import (
	"testing"
)

func Test_ServiceIDExtractor(t *testing.T) {
	extractor := &ServiceIDExtractor{
		ServiceAliases: map[string]string{
			"eth": "F00C",
		},
	}

	tests := []struct {
		name    string
		headers map[string]string
		host    string
		want    string
		wantErr bool
	}{
		{
			name: "should extract service ID from target-service-id header",
			headers: map[string]string{
				reqHeaderServiceID: "eth",
			},
			host:    "example.com",
			want:    "F00C",
			wantErr: false,
		},
		{
			name:    "should extract service ID from subdomain",
			headers: map[string]string{},
			host:    "eth.example.com",
			want:    "F00C",
			wantErr: false,
		},
		{
			name:    "should return error if service ID is not found",
			headers: map[string]string{},
			host:    "example.com",
			want:    "",
			wantErr: true,
		},
		{
			name: "should return service ID directly if not an alias",
			headers: map[string]string{
				reqHeaderServiceID: "non-alias-id",
			},
			host:    "example.com",
			want:    "non-alias-id",
			wantErr: false,
		},
		{
			name:    "should return error if both header and subdomain are empty",
			headers: map[string]string{},
			host:    "",
			want:    "",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := extractor.extractServiceID(test.headers, test.host)
			if (err != nil) != test.wantErr {
				t.Errorf("extractServiceID() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if got != test.want {
				t.Errorf("extractServiceID() = %v, want %v", got, test.want)
			}
		})
	}
}

package shannon

// Test TLD extraction functionality (focused test)
// func TestTLDExtractionLogic(t *testing.T) {
// 	tests := []struct {
// 		name         string
// 		endpointAddr string
// 		expectedTLD  string
// 	}{
// 		{
// 			name:         "com_domain",
// 			endpointAddr: "supplier1-https://api.example.com/v1",
// 			expectedTLD:  "example.com",
// 		},
// 		{
// 			name:         "org_domain",
// 			endpointAddr: "supplier2-https://api.example.org:8080",
// 			expectedTLD:  "example.org",
// 		},
// 		{
// 			name:         "io_domain",
// 			endpointAddr: "supplier3-api.example.io",
// 			expectedTLD:  "example.io",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tld, err := shannonmetrics.ExtractEffectiveTLDPlusOne(tt.endpointAddr)
// 			require.NoError(t, err)
// 			require.Equal(t, tt.expectedTLD, tld, "TLD extraction failed for: %s", tt.endpointAddr)
// 		})
// 	}
// }

// Test TLD extraction functionality
// func TestExtractTLDFromEndpointAddr(t *testing.T) {
// 	tests := []struct {
// 		name         string
// 		endpointAddr string
// 		expectedTLD  string
// 	}{
// 		{
// 			name:         "standard_url",
// 			endpointAddr: "supplier1-https://api.example.com/v1",
// 			expectedTLD:  "example.com",
// 		},
// 		{
// 			name:         "url_with_port",
// 			endpointAddr: "supplier2-https://api.example.net:8080",
// 			expectedTLD:  "example.net",
// 		},
// 		{
// 			name:         "encoded_url",
// 			endpointAddr: "supplier3-https%3A%2F%2Fapi.example.org",
// 			expectedTLD:  "example.org",
// 		},
// 		{
// 			name:         "no_protocol",
// 			endpointAddr: "supplier4-api.example.io",
// 			expectedTLD:  "example.io",
// 		},
// 		{
// 			name:         "localhost",
// 			endpointAddr: "supplier5-http://localhost:8080",
// 			expectedTLD:  "localhost",
// 		},
// 		{
// 			name:         "ip_address",
// 			endpointAddr: "supplier6-http://192.168.1.1:8080",
// 			expectedTLD:  "1.1", // Current behavior: extracts last part of IP as TLD
// 		},
// 		{
// 			name:         "malformed_url",
// 			endpointAddr: "invalid-endpoint",
// 			expectedTLD:  "",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tld, err := shannonmetrics.ExtractEffectiveTLDPlusOne(tt.endpointAddr)
// 			require.NoError(t, err)
// 			require.Equal(t, tt.expectedTLD, tld, "TLD extraction failed for: %s", tt.endpointAddr)
// 		})
// 	}
// }

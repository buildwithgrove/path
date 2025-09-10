package protocol

import (
	"testing"
)

func TestEndpointAddr_GetURL(t *testing.T) {
	tests := []struct {
		name         string
		endpointAddr EndpointAddr
		expected     string
		expectError  bool
	}{
		// Standard endpoint addresses
		{
			name:         "NodeFleet endpoint",
			endpointAddr: EndpointAddr("pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net"),
			expected:     "https://relayminer.shannon-mainnet.eu.nodefleet.net",
		},
		{
			name:         "DoPokT endpoint with port",
			endpointAddr: EndpointAddr("pokt1d3atlnepcvsa9j5uunpvf64g80eucjqtem77mn-https://rm02-eu.dopokt.com:443"),
			expected:     "https://rm02-eu.dopokt.com:443",
		},
		{
			name:         "Simple HTTPS endpoint",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://example.com"),
			expected:     "https://example.com",
		},
		{
			name:         "HTTP endpoint with port",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-http://api.example.com:8080"),
			expected:     "http://api.example.com:8080",
		},

		// URLs with dashes (testing dash handling)
		{
			name:         "URL with dashes in hostname",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://my-service.example.com"),
			expected:     "https://my-service.example.com",
		},
		{
			name:         "URL with dashes in path",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://api.example.com/v1/some-endpoint"),
			expected:     "https://api.example.com/v1/some-endpoint",
		},
		{
			name:         "Complex URL with many dashes",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://multi-dash-service.sub-domain.example-site.com:8080/api/v1/test-endpoint"),
			expected:     "https://multi-dash-service.sub-domain.example-site.com:8080/api/v1/test-endpoint",
		},

		// IP address endpoints
		{
			name:         "IPv4 address with HTTPS",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://192.168.1.100:8545"),
			expected:     "https://192.168.1.100:8545",
		},
		{
			name:         "IPv4 address with HTTP",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-http://10.0.0.1:3000"),
			expected:     "http://10.0.0.1:3000",
		},
		{
			name:         "IPv6 address",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://[::1]:8080"),
			expected:     "https://[::1]:8080",
		},
		{
			name:         "IPv6 address with zone",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://[fe80::1%lo0]:8080"),
			expected:     "https://[fe80::1%lo0]:8080",
		},
		{
			name:         "Public IPv4",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://203.0.113.42:8545"),
			expected:     "https://203.0.113.42:8545",
		},

		// Localhost endpoints
		{
			name:         "Localhost with HTTPS",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://localhost:3000"),
			expected:     "https://localhost:3000",
		},
		{
			name:         "Localhost with HTTP",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-http://localhost:8080"),
			expected:     "http://localhost:8080",
		},
		{
			name:         "Localhost.localdomain",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-http://localhost.localdomain:3000"),
			expected:     "http://localhost.localdomain:3000",
		},
		{
			name:         "Localhost subdomain",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://api.localhost.test"),
			expected:     "https://api.localhost.test",
		},

		// Internal/private domain endpoints
		{
			name:         "Internal domain",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://relay.company.internal:8545"),
			expected:     "https://relay.company.internal:8545",
		},
		{
			name:         "Local domain",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://service.local"),
			expected:     "https://service.local",
		},
		{
			name:         "Corp domain",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://blockchain.corp:9545"),
			expected:     "https://blockchain.corp:9545",
		},
		{
			name:         "Home domain",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://pi-node.home:8545"),
			expected:     "https://pi-node.home:8545",
		},
		{
			name:         "LAN domain",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://router.lan"),
			expected:     "https://router.lan",
		},

		// Single hostname endpoints
		{
			name:         "Single hostname with HTTPS",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://relayminer1"),
			expected:     "https://relayminer1",
		},
		{
			name:         "Single hostname with HTTP",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-http://server1:8080"),
			expected:     "http://server1:8080",
		},
		{
			name:         "Single hostname uppercase",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://RELAYMINER1"),
			expected:     "https://RELAYMINER1",
		},

		// URLs with paths and query parameters
		{
			name:         "URL with path",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://api.example.com/v1/rpc"),
			expected:     "https://api.example.com/v1/rpc",
		},
		{
			name:         "URL with query parameters",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://api.example.com/rpc?version=1&format=json"),
			expected:     "https://api.example.com/rpc?version=1&format=json",
		},
		{
			name:         "URL with fragment",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://api.example.com/docs#section1"),
			expected:     "https://api.example.com/docs#section1",
		},
		{
			name:         "URL with auth",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://user:pass@api.example.com"),
			expected:     "https://user:pass@api.example.com",
		},
		{
			name:         "Complex URL with everything",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://user:pass@api.example.com:8080/v1/rpc?key=value&format=json#docs"),
			expected:     "https://user:pass@api.example.com:8080/v1/rpc?key=value&format=json#docs",
		},

		// Different address formats
		{
			name:         "Short address",
			endpointAddr: EndpointAddr("pokt123-https://example.com"),
			expected:     "https://example.com",
		},
		{
			name:         "Long address",
			endpointAddr: EndpointAddr("pokt1234567890abcdef1234567890abcdef1234567890abcdef-https://example.com"),
			expected:     "https://example.com",
		},
		{
			name:         "Address with dashes",
			endpointAddr: EndpointAddr("pokt-test-address-123-https://example.com"),
			expected:     "test-address-123-https://example.com",
		},

		// Real-world examples
		{
			name:         "Real NodeFleet endpoint",
			endpointAddr: EndpointAddr("pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net"),
			expected:     "https://relayminer.shannon-mainnet.eu.nodefleet.net",
		},
		{
			name:         "Real DoPokT endpoint",
			endpointAddr: EndpointAddr("pokt1d3atlnepcvsa9j5uunpvf64g80eucjqtem77mn-https://rm02-eu.dopokt.com:443"),
			expected:     "https://rm02-eu.dopokt.com:443",
		},

		// Edge cases
		{
			name:         "Minimal valid endpoint",
			endpointAddr: EndpointAddr("a-b"),
			expected:     "b",
		},
		{
			name:         "URL with encoded characters",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://api.example.com/path%20with%20spaces"),
			expected:     "https://api.example.com/path%20with%20spaces",
		},

		// Error cases
		{
			name:         "Missing dash separator",
			endpointAddr: EndpointAddr("pokt1234567890abcdefhttps://example.com"),
			expectError:  true,
		},
		{
			name:         "Empty URL part",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-"),
			expectError:  true,
		},
		{
			name:         "Only dash",
			endpointAddr: EndpointAddr("-"),
			expectError:  true,
		},
		{
			name:         "Empty endpoint address",
			endpointAddr: EndpointAddr(""),
			expectError:  true,
		},
		{
			name:         "No address part",
			endpointAddr: EndpointAddr("-https://example.com"),
			expected:     "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.endpointAddr.GetURL()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEndpointAddr_GetAddress(t *testing.T) {
	tests := []struct {
		name         string
		endpointAddr EndpointAddr
		expected     string
		expectError  bool
	}{
		// Standard endpoint addresses
		{
			name:         "Standard endpoint",
			endpointAddr: EndpointAddr("pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net"),
			expected:     "pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq",
		},
		{
			name:         "DoPokT endpoint",
			endpointAddr: EndpointAddr("pokt1d3atlnepcvsa9j5uunpvf64g80eucjqtem77mn-https://rm02-eu.dopokt.com:443"),
			expected:     "pokt1d3atlnepcvsa9j5uunpvf64g80eucjqtem77mn",
		},
		{
			name:         "Simple address",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://example.com"),
			expected:     "pokt1234567890abcdef",
		},

		// URLs with dashes (testing first dash extraction)
		{
			name:         "URL with dashes in hostname",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://my-service.example.com"),
			expected:     "pokt1234567890abcdef",
		},
		{
			name:         "URL with dashes in path",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://api.example.com/v1/some-endpoint"),
			expected:     "pokt1234567890abcdef",
		},
		{
			name:         "Complex URL with many dashes",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://multi-dash-service.sub-domain.example-site.com:8080/api/v1/test-endpoint"),
			expected:     "pokt1234567890abcdef",
		},

		// Different address formats
		{
			name:         "Short address",
			endpointAddr: EndpointAddr("pokt123-https://example.com"),
			expected:     "pokt123",
		},
		{
			name:         "Long address",
			endpointAddr: EndpointAddr("pokt1234567890abcdef1234567890abcdef1234567890abcdef-https://example.com"),
			expected:     "pokt1234567890abcdef1234567890abcdef1234567890abcdef",
		},
		{
			name:         "Address with dashes - should extract before first dash",
			endpointAddr: EndpointAddr("pokt-test-address-123-https://example.com"),
			expected:     "pokt",
		},
		{
			name:         "Numeric address",
			endpointAddr: EndpointAddr("123456789-https://example.com"),
			expected:     "123456789",
		},
		{
			name:         "Mixed case address",
			endpointAddr: EndpointAddr("PoKt1AbCdEf-https://example.com"),
			expected:     "PoKt1AbCdEf",
		},

		// Different URL types
		{
			name:         "IP address URL",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://192.168.1.100:8545"),
			expected:     "pokt1234567890abcdef",
		},
		{
			name:         "Localhost URL",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-http://localhost:3000"),
			expected:     "pokt1234567890abcdef",
		},
		{
			name:         "Internal domain URL",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://relay.company.internal:8545"),
			expected:     "pokt1234567890abcdef",
		},
		{
			name:         "Single hostname URL",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://relayminer1"),
			expected:     "pokt1234567890abcdef",
		},

		// URLs with paths and parameters
		{
			name:         "URL with path",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://api.example.com/v1/rpc"),
			expected:     "pokt1234567890abcdef",
		},
		{
			name:         "URL with query parameters",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://api.example.com/rpc?version=1&format=json"),
			expected:     "pokt1234567890abcdef",
		},
		{
			name:         "URL with fragment",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://api.example.com/docs#section1"),
			expected:     "pokt1234567890abcdef",
		},
		{
			name:         "URL with auth",
			endpointAddr: EndpointAddr("pokt1234567890abcdef-https://user:pass@api.example.com"),
			expected:     "pokt1234567890abcdef",
		},

		// Real-world examples
		{
			name:         "Real NodeFleet address",
			endpointAddr: EndpointAddr("pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net"),
			expected:     "pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq",
		},
		{
			name:         "Real DoPokT address",
			endpointAddr: EndpointAddr("pokt1d3atlnepcvsa9j5uunpvf64g80eucjqtem77mn-https://rm02-eu.dopokt.com:443"),
			expected:     "pokt1d3atlnepcvsa9j5uunpvf64g80eucjqtem77mn",
		},

		// Edge cases
		{
			name:         "Minimal valid endpoint",
			endpointAddr: EndpointAddr("a-b"),
			expected:     "a",
		},
		{
			name:         "Single character address",
			endpointAddr: EndpointAddr("x-https://example.com"),
			expected:     "x",
		},
		{
			name:         "Empty address part",
			endpointAddr: EndpointAddr("-https://example.com"),
			expected:     "",
		},
		{
			name:         "Address with special characters",
			endpointAddr: EndpointAddr("addr_123!@#-https://example.com"),
			expected:     "addr_123!@#",
		},

		// Error cases
		{
			name:         "Missing dash separator",
			endpointAddr: EndpointAddr("pokt1234567890abcdefhttps://example.com"),
			expectError:  true,
		},
		{
			name:         "Only dash",
			endpointAddr: EndpointAddr("-"),
			expected:     "",
		},
		{
			name:         "Empty endpoint address",
			endpointAddr: EndpointAddr(""),
			expectError:  true,
		},
		{
			name:         "No dash at all",
			endpointAddr: EndpointAddr("pokt1234567890abcdef"),
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.endpointAddr.GetAddress()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEndpointAddr_String(t *testing.T) {
	tests := []struct {
		name         string
		endpointAddr EndpointAddr
		expected     string
	}{
		{
			name:         "Standard endpoint",
			endpointAddr: EndpointAddr("pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net"),
			expected:     "pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net",
		},
		{
			name:         "Simple endpoint",
			endpointAddr: EndpointAddr("address-url"),
			expected:     "address-url",
		},
		{
			name:         "Empty endpoint",
			endpointAddr: EndpointAddr(""),
			expected:     "",
		},
		{
			name:         "Special characters",
			endpointAddr: EndpointAddr("addr!@#$%^&*()-https://example.com/path?query=value#fragment"),
			expected:     "addr!@#$%^&*()-https://example.com/path?query=value#fragment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.endpointAddr.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestEndpointAddrList_String(t *testing.T) {
	tests := []struct {
		name     string
		list     EndpointAddrList
		expected string
	}{
		{
			name: "Multiple endpoints",
			list: EndpointAddrList{
				EndpointAddr("pokt1addr1-https://example1.com"),
				EndpointAddr("pokt1addr2-https://example2.com"),
				EndpointAddr("pokt1addr3-https://example3.com"),
			},
			expected: "pokt1addr1-https://example1.com, pokt1addr2-https://example2.com, pokt1addr3-https://example3.com",
		},
		{
			name: "Single endpoint",
			list: EndpointAddrList{
				EndpointAddr("pokt1addr1-https://example1.com"),
			},
			expected: "pokt1addr1-https://example1.com",
		},
		{
			name:     "Empty list",
			list:     EndpointAddrList{},
			expected: "",
		},
		{
			name: "Real-world endpoints",
			list: EndpointAddrList{
				EndpointAddr("pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net"),
				EndpointAddr("pokt1d3atlnepcvsa9j5uunpvf64g80eucjqtem77mn-https://rm02-eu.dopokt.com:443"),
			},
			expected: "pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net, pokt1d3atlnepcvsa9j5uunpvf64g80eucjqtem77mn-https://rm02-eu.dopokt.com:443",
		},
		{
			name: "Endpoints with various URL types",
			list: EndpointAddrList{
				EndpointAddr("addr1-https://192.168.1.1:8545"),
				EndpointAddr("addr2-http://localhost:3000"),
				EndpointAddr("addr3-https://service.local"),
				EndpointAddr("addr4-https://relayminer1"),
			},
			expected: "addr1-https://192.168.1.1:8545, addr2-http://localhost:3000, addr3-https://service.local, addr4-https://relayminer1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.list.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Integration tests combining GetAddress and GetURL
func TestEndpointAddr_Integration(t *testing.T) {
	tests := []struct {
		name            string
		endpointAddr    EndpointAddr
		expectedAddress string
		expectedURL     string
		expectError     bool
	}{
		{
			name:            "Standard endpoint integration",
			endpointAddr:    EndpointAddr("pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net"),
			expectedAddress: "pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq",
			expectedURL:     "https://relayminer.shannon-mainnet.eu.nodefleet.net",
		},
		{
			name:            "URL with dashes integration",
			endpointAddr:    EndpointAddr("pokt1d3atlnepcvsa9j5uunpvf64g80eucjqtem77mn-https://rm02-eu.dopokt.com:443"),
			expectedAddress: "pokt1d3atlnepcvsa9j5uunpvf64g80eucjqtem77mn",
			expectedURL:     "https://rm02-eu.dopokt.com:443",
		},
		{
			name:            "Complex URL with many dashes",
			endpointAddr:    EndpointAddr("address-with-dashes-https://multi-dash-service.sub-domain.example-site.com:8080/api/v1/test-endpoint"),
			expectedAddress: "address",
			expectedURL:     "with-dashes-https://multi-dash-service.sub-domain.example-site.com:8080/api/v1/test-endpoint",
		},
		{
			name:         "Invalid endpoint - no dash",
			endpointAddr: EndpointAddr("invalidendpoint"),
			expectError:  true,
		},
		{
			name:         "Invalid endpoint - empty",
			endpointAddr: EndpointAddr(""),
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, addrErr := tt.endpointAddr.GetAddress()
			url, urlErr := tt.endpointAddr.GetURL()

			if tt.expectError {
				if addrErr == nil || urlErr == nil {
					t.Errorf("Expected error for both GetAddress and GetURL but got: addrErr=%v, urlErr=%v", addrErr, urlErr)
				}
				return
			}

			if addrErr != nil {
				t.Errorf("Unexpected error from GetAddress: %v", addrErr)
				return
			}

			if urlErr != nil {
				t.Errorf("Unexpected error from GetURL: %v", urlErr)
				return
			}

			if address != tt.expectedAddress {
				t.Errorf("GetAddress: expected %q, got %q", tt.expectedAddress, address)
			}

			if url != tt.expectedURL {
				t.Errorf("GetURL: expected %q, got %q", tt.expectedURL, url)
			}

			// Verify that combining address and URL gives back original endpoint
			reconstructed := address + "-" + url
			if reconstructed != string(tt.endpointAddr) {
				t.Errorf("Reconstruction failed: expected %q, got %q", string(tt.endpointAddr), reconstructed)
			}
		})
	}
}

// Benchmark tests
func BenchmarkEndpointAddr_GetAddress(b *testing.B) {
	endpoint := EndpointAddr("pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = endpoint.GetAddress()
	}
}

func BenchmarkEndpointAddr_GetURL(b *testing.B) {
	endpoint := EndpointAddr("pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = endpoint.GetURL()
	}
}

func BenchmarkEndpointAddrList_String(b *testing.B) {
	list := EndpointAddrList{
		EndpointAddr("pokt1eetcwfv2agdl2nvpf4cprhe89rdq3cxdf037wq-https://relayminer.shannon-mainnet.eu.nodefleet.net"),
		EndpointAddr("pokt1d3atlnepcvsa9j5uunpvf64g80eucjqtem77mn-https://rm02-eu.dopokt.com:443"),
		EndpointAddr("pokt1234567890abcdef-https://192.168.1.1:8545"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = list.String()
	}
}

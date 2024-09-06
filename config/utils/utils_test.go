package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_IsValidHex(t *testing.T) {
	test := []struct {
		name   string
		input  string
		length int
		want   bool
	}{
		{
			name:   "should return true for valid hex of correct length",
			input:  "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf619",
			length: 64,
			want:   true,
		},
		{
			name:   "should return false for valid hex of incorrect length",
			input:  "a6258b46ecad0628b72099f91e87eef1b040a87",
			length: 64,
			want:   false,
		},
		{
			name:   "should return false for invalid hex of correct length",
			input:  "a6258b46ecad0628b72099f91e87eef1b040a8747ed2d476f56ad359372bf61y",
			length: 64,
			want:   false,
		},
		{
			name:   "should return false for invalid hex of incorrect length",
			input:  "a6258b46ecad0628b72099f91e87eef1b04y",
			length: 64,
			want:   false,
		},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			got := IsValidHex(test.input, test.length)
			c.Equal(test.want, got)
		})
	}
}

func Test_IsValidSubdomain(t *testing.T) {
	test := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "should return true for valid subdomain",
			input: "valid-subdomain",
			want:  true,
		},
		{
			name:  "should return false for subdomain with invalid characters",
			input: "invalid_subdomain!",
			want:  false,
		},
		{
			name:  "should return false for subdomain starting with hyphen",
			input: "-invalid",
			want:  false,
		},
		{
			name:  "should return false for subdomain ending with hyphen",
			input: "invalid-",
			want:  false,
		},
		{
			name:  "should return true for subdomain with numbers",
			input: "subdomain123",
			want:  true,
		},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			got := IsValidSubdomain(test.input)
			c.Equal(test.want, got)
		})
	}
}
func Test_IsValidURL(t *testing.T) {
	test := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "should return true for valid http URL",
			input: "http://example.com",
			want:  true,
		},
		{
			name:  "should return true for valid https URL",
			input: "https://example.com",
			want:  true,
		},
		{
			name:  "should return false for invalid URL",
			input: "htp://example.com",
			want:  false,
		},
		{
			name:  "should return false for URL with spaces",
			input: "http://example .com",
			want:  false,
		},
		{
			name:  "should return false for URL with invalid scheme",
			input: "example.com",
			want:  false,
		},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			got := IsValidURL(test.input)
			c.Equal(test.want, got)
		})
	}
}

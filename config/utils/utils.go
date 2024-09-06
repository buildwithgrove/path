package utils

import (
	"encoding/hex"
	"net/url"
	"regexp"
)

// IsValidHex checks if a string is a valid hex code of a given length.
func IsValidHex(s string, length int) bool {
	if len(s) != length {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

// IsValidSubdomain checks if a string is a valid URL subdomain.
func IsValidSubdomain(s string) bool {
	// Regular expression to match valid subdomains
	var subdomainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
	return subdomainRegex.MatchString(s)
}

// IsValidURL checks if a string is a valid URL.
func IsValidURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// IsValidDBConnectionString checks if a string is a valid PostgreSQL connection string.
func IsValidDBConnectionString(s string) bool {
	// Regular expression to match a valid PostgreSQL connection string
	var dbConnStringRegex = regexp.MustCompile(`^postgres://[^:]+:[^@]+@[^:]+:\d+/.+$`)
	return dbConnStringRegex.MatchString(s)
}

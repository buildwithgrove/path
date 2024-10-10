//go:build auth_server

package server

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/buildwithgrove/auth-server/user"
	auth_pb "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/golang-jwt/jwt/v5"
)

// JWTParser handles JWT parsing and validation
type JWTParser struct {
	Issuer   string
	Audience string
	JWKSURL  string

	mu       sync.Mutex
	keyCache map[string]*rsa.PublicKey
}

// ParseJWT parses and validates the JWT from the request
func (a *JWTParser) ParseJWT(req *auth_pb.CheckRequest) (user.EndpointID, *errorResponse) {
	// Extract HTTP request attributes
	httpReq := req.GetAttributes().GetRequest().GetHttp()
	headers := httpReq.GetHeaders()

	// Get the Authorization header
	authHeader, ok := headers["authorization"]
	if !ok || authHeader == "" {
		return "", &errAuthorizationHeaderRequired
	}

	// Validate the Authorization header format
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", &errInvalidAuthorizationHeader
	}

	// Extract the JWT token
	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))

	// Parse and validate the JWT token
	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token is signed using RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID (kid) from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}

		// Fetch the RSA public key
		return a.getKey(kid)
	})
	if err != nil || !token.Valid {
		return "", &errInvalidToken
	}

	// Manually validate standard claims
	if claims.Issuer != a.Issuer {
		return "", &errInvalidIssuer
	}
	if !contains(claims.Audience, a.Audience) {
		return "", &errInvalidAudience
	}
	if claims.ExpiresAt.Before(time.Now()) {
		return "", &errTokenExpired
	}

	// Extract endpoint ID from token claims
	endpointID := claims.EndpointID
	if endpointID == "" {
		return "", &errEndpointIDNotFound
	}

	// Authorization succeeded, return endpoint ID
	return user.EndpointID(endpointID), nil
}

// contains checks if a string slice contains a specific string
func contains(audiences jwt.ClaimStrings, audience string) bool {
	for _, aud := range audiences {
		if aud == audience {
			return true
		}
	}
	return false
}

// CustomClaims represents the JWT claims with custom fields
type CustomClaims struct {
	EndpointID string `json:"endpoint_id"`
	jwt.RegisteredClaims
}

// getKey retrieves the RSA public key for the given key ID (kid)
func (a *JWTParser) getKey(kid string) (*rsa.PublicKey, error) {
	// Check if the key is cached
	a.mu.Lock()
	if a.keyCache == nil {
		a.keyCache = make(map[string]*rsa.PublicKey)
	}
	key, exists := a.keyCache[kid]
	a.mu.Unlock()
	if exists {
		return key, nil
	}

	// Fetch the JWKS
	resp, err := http.Get(a.JWKSURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %v", err)
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %v", err)
	}

	// Find the key with the matching kid
	for _, jwk := range jwks.Keys {
		if jwk.Kid == kid {
			// Parse the RSA public key
			pubKey, err := jwk.ParsePublicKey()
			if err != nil {
				return nil, fmt.Errorf("failed to parse public key: %v", err)
			}

			// Cache the key for future use
			a.mu.Lock()
			a.keyCache[kid] = pubKey
			a.mu.Unlock()

			return pubKey, nil
		}
	}

	return nil, fmt.Errorf("unable to find key with kid: %s", kid)
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"` // Modulus
	E   string `json:"e"` // Exponent
}

// ParsePublicKey parses the JWK and returns an RSA public key
func (jwk *JWK) ParsePublicKey() (*rsa.PublicKey, error) {
	if jwk.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported key type %s", jwk.Kty)
	}

	// Decode the modulus and exponent
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %v", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E: %v", err)
	}

	// Convert exponent bytes to integer
	var eInt int
	for _, b := range eBytes {
		eInt = eInt<<8 + int(b)
	}

	// Construct the RSA public key
	pubKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eInt,
	}
	return pubKey, nil
}

// Error responses
var (
	errAuthorizationHeaderRequired = errorResponse{
		code:    http.StatusUnauthorized,
		message: "Authorization header required",
	}
	errInvalidAuthorizationHeader = errorResponse{
		code:    http.StatusUnauthorized,
		message: "Invalid Authorization header format",
	}
	errInvalidToken = errorResponse{
		code:    http.StatusUnauthorized,
		message: "Invalid or malformed JWT",
	}
	errTokenExpired = errorResponse{
		code:    http.StatusUnauthorized,
		message: "JWT has expired",
	}
	errInvalidIssuer = errorResponse{
		code:    http.StatusUnauthorized,
		message: "Invalid token issuer",
	}
	errInvalidAudience = errorResponse{
		code:    http.StatusUnauthorized,
		message: "Invalid token audience",
	}
	errEndpointIDNotFound = errorResponse{
		code:    http.StatusUnauthorized,
		message: "Endpoint ID not found in token claims",
	}
)

// errorResponse represents an HTTP error response
type errorResponse struct {
	code    int
	message string
}

package cometbft

// requestContextFromInternalError creates a request context for internal errors.
// Example: errors reading HTTP request body.
func requestContextFromInternalError(_ error) *requestContext {
	// TODO_MVP(@adshmh): return a request context to handle internal errors.
	return nil
}

// requestContextFromUserError creates a request context for user errors.
// // Example: invalid JSON-RPC request body that cannot be unmarshalled).
func requestContextFromUserError(_ error) *requestContext {
	// TODO_MVP(@adshmh): return a request context to handle user errors.
	return nil
}

package cometbft

// TODO_MVP(@adshmh): return a request context to handle internal errors.
// requestContextFromInternalError returns a request context
// for an internal error, e.g. error on reading the HTTP request body.
func requestContextFromInternalError(err error) *requestContext {
	return nil
}

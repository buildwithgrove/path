package cometbft

// TODO_MVP(@adshmh): return a request context to handle internal errors.
// requestContextFromInternalError returns a request context
// for an internal error, e.g. error on reading the HTTP request body.
func requestContextFromInternalError(_ error) *requestContext {
	return nil
}

// TODO_MVP(@adshmh): return a request context to handle user errors.
// requestContextFromUserError returns a request context
// for a user error, e.g. an unmarshalling error is a
// user error because the request body, provided by the user,
// cannot be parsed as a valid JSONRPC request.
func requestContextFromUserError(_ error) *requestContext {
	return nil
}

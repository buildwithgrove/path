package gateway

// HTTPResponse allows passing around an HTTP response intended to be received
// by the user. It is used instead of the standard http.Response to minimze the
// requirements for producers and clarify the data items that will be used when
// the writing of the HTTP response takes place.
type HTTPResponse interface {
	GetPayload() []byte
	GetHTTPStatusCode() int
	// TODO_IMPROVE: return http.Header instead of a map[string]string.
	GetHTTPHeaders() map[string]string
}

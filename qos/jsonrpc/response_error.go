package jsonrpc

// ResponseError captures a JSONRPC response error struct
// See the following link for more details:
// https://www.jsonrpc.org/specification#error_object
type ResponseError struct {
	// A Number that indicates the error type that occurred.
	Code int `json:"code"`
	// A String providing a short description of the error.
	Message string `json:"message"`
	// A Primitive or Structured value that contains additional information about the error.
	// This may be omitted.
	// The value of this member is defined by the Server (e.g. detailed error information, nested errors etc.).
	Data any `json:"data,omitempty"`
}

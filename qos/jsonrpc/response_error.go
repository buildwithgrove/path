package jsonrpc

// ResponseError captures a JSONRPC response error struct
// See the following link for more details:
// https://www.jsonrpc.org/specification#error_object
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

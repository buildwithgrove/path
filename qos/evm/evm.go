package evm

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/relayer"
)

const (
	defaultEVMRequestTimeout = 5 * time.Second
	// TODO_IMPROVE: return the same request ID as the request that caused the error
	evmErrorTemplate = `{"jsonrpc":"2.0","id":"0","error":{"code":-32603,"message":"%s"}}`
)

var ( // compile-time checks to ensure EVMServiceQoS implements the required interfaces
	_ gateway.QoSService         = &EVMServiceQoS{}
	_ gateway.QoSResponseBuilder = &EVMResponseBuilder{}
	_ gateway.QoSRequestParser   = &EVMRequestParser{}
	_ gateway.HTTPResponse       = &EVMHTTPResponse{}
)

// EVMServiceQoS is the QoS service for EVM-based chains, which handles logic specific
// to EVM-based chains, such as request parsing, response building, and endpoint selection.
type EVMServiceQoS struct {
	EVMRequestParser
	EVMResponseBuilder
	EVMEndpointSelector
}

func NewEVMServiceQoS(requestTimeout time.Duration, logger polylog.Logger) *EVMServiceQoS {
	return &EVMServiceQoS{
		EVMRequestParser: EVMRequestParser{
			requestTimeout: requestTimeout,
			logger:         logger,
		},
		EVMResponseBuilder:  EVMResponseBuilder{},
		EVMEndpointSelector: EVMEndpointSelector{},
	}
}

/* EVM QoS Request Parser */

type EVMRequestParser struct {
	requestTimeout time.Duration
	logger         polylog.Logger
}

func (p *EVMRequestParser) ParseHTTPRequest(ctx context.Context, req *http.Request) (relayer.Payload, error) {

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return relayer.Payload{}, err
	}

	if p.requestTimeout == 0 {
		p.requestTimeout = defaultEVMRequestTimeout
	}

	return relayer.Payload{
		Data:            string(body),
		Method:          req.Method,
		Path:            req.URL.Path,
		TimeoutMillisec: int(p.requestTimeout.Milliseconds()),
	}, nil
}

/* EVM QoS Response Builder */

type EVMResponseBuilder struct{}

func (b *EVMResponseBuilder) GetHTTPResponse(ctx context.Context, resp relayer.Response) (gateway.HTTPResponse, error) {
	return &EVMHTTPResponse{
		payload:        resp.Bytes,
		httpStatusCode: resp.HTTPStatusCode,
		httpHeaders:    http.Header{},
	}, nil
}

// TODO_INCOMPLETE: This method needs to validate the response is a valid JSON-RPC object
func (b *EVMResponseBuilder) GetHTTPErrorResponse(ctx context.Context, err error) gateway.HTTPResponse {

	return &EVMHTTPResponse{
		payload: []byte(fmt.Sprintf(evmErrorTemplate, err.Error())),
		// TODO_TECHDEBT: get HTTP status code from specific error type
		httpStatusCode: http.StatusInternalServerError,
		httpHeaders:    http.Header{},
	}
}

/* EVM Endpoint Selector */

// TODO_INCOMPLETE: implement a proper endpoint selector for EVM chains
type EVMEndpointSelector struct{}

// TODO_INCOMPLETE: implement actual logic for selecting the most appropriate endpoint using QoS metrics
func (s *EVMEndpointSelector) Select(endpoints map[relayer.AppAddr][]relayer.Endpoint) (relayer.AppAddr, relayer.EndpointAddr, error) {
	for appAddr, endpointList := range endpoints {
		if len(endpointList) > 0 {
			randomIndex := rand.Intn(len(endpointList))
			return appAddr, endpointList[randomIndex].Addr(), nil
		}
	}
	return "", "", fmt.Errorf("no endpoints available")
}

/* EVM HTTP Response */

type EVMHTTPResponse struct {
	payload        []byte
	httpStatusCode int
	httpHeaders    http.Header
}

func (r *EVMHTTPResponse) GetPayload() []byte {
	return r.payload
}

func (r *EVMHTTPResponse) GetHTTPStatusCode() int {
	return r.httpStatusCode
}

func (r *EVMHTTPResponse) GetHTTPHeaders() map[string]string {
	headersMap := make(map[string]string)
	for key, value := range r.httpHeaders {
		headersMap[key] = value[0]
	}
	return headersMap
}

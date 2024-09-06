package shannon

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	servicetypes "github.com/pokt-network/poktroll/x/service/types"
)

// sendHttpRelay sends the relay request to the supplier at the given URL using an HTTP Post request.
func sendHttpRelay(
	ctx context.Context,
	supplierUrlStr string,
	relayRequest *servicetypes.RelayRequest,
) (relayResponseBz []byte, err error) {
	supplierUrl, err := url.Parse(supplierUrlStr)
	if err != nil {
		return nil, err
	}

	relayRequestBz, err := relayRequest.Marshal()
	if err != nil {
		return nil, err
	}

	relayHTTPRequest := &http.Request{
		Method: http.MethodPost,
		URL:    supplierUrl,
		Body:   io.NopCloser(bytes.NewReader(relayRequestBz)),
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	relayHTTPResponse, err := http.DefaultClient.Do(relayHTTPRequest)
	if err != nil {
		return nil, err
	}
	defer relayHTTPResponse.Body.Close()

	return io.ReadAll(relayHTTPResponse.Body)
}

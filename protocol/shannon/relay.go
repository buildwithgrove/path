package shannon

import (
	"bytes"
	"context"
	"fmt"
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
	headers map[string]string,
) (relayResponseBz []byte, err error) {
	_, err = url.Parse(supplierUrlStr)
	if err != nil {
		return nil, err
	}

	relayRequestBz, err := relayRequest.Marshal()
	if err != nil {
		return nil, err
	}

	relayHTTPRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		supplierUrlStr,
		io.NopCloser(bytes.NewReader(relayRequestBz)),
	)
	if err != nil {
		return nil, err
	}

	relayHTTPRequest.Header.Add("Content-Type", "application/json")
	for key, value := range headers {
		relayHTTPRequest.Header.Add(key, value)
	}

	fmt.Println("DEBUG - relayHTTPRequest Headers", relayHTTPRequest.Header)

	// TODO_IMPROVE(@commoddity): Use a custom HTTP client to:
	//  - allow configuring the defaultTransport.
	//  - allow PATH users to override default transport config.
	//
	// Best practice in Go is to use a custom HTTP client Transport.
	// See: https://vishnubharathi.codes/blog/know-when-to-break-up-with-go-http-defaultclient/
	relayHTTPResponse, err := http.DefaultClient.Do(relayHTTPRequest)
	if err != nil {
		return nil, err
	}
	defer relayHTTPResponse.Body.Close()

	return io.ReadAll(relayHTTPResponse.Body)
}

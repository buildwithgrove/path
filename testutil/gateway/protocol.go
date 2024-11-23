package gateway 

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/protocol"
	gatewaymocks"github.com/buildwithgrove/path/testutil/gateway/mocks"
	protocolmocks "github.com/buildwithgrove/path/testutil/protocol/mocks"
)

// ProtocolOpt is a function which receives and potentially modifies protocol instances during construction.
//type ProtocolOpt func(....) 

func endpoint(ctrl *gomock.Controller, addr, url string) protocol.Endpoint {
	ep := protocolmocks.NewMockEndpoint(ctrl)
	ep.EXPECT().Addr().Return(protocol.EndpointAddr(addr)).AnyTimes()
	ep.EXPECT().PublicURL().Return(url).AnyTimes()

	return ep
}

func endpoints(ctrl *gomock.Controller) []protocol.Endpoint {
	return []protocol.Endpoint{
		endpoint(ctrl, "addr1", "url1"),
		endpoint(ctrl, "addr2", "url2"),
		endpoint(ctrl, "addr3", "url3"),
		endpoint(ctrl, "addr4", "url4"),
	}
}

func Protocol(t *testing.T) gateway.Protocol {
	t.Helper()

	ctrl := gomock.NewController(t)

	mockReqCtx := gatewaymocks.NewMockProtocolRequestContext(ctrl)

	availableEndpoints := endpoints(ctrl)

	mockReqCtx.EXPECT().
		AvailableEndpoints().
		DoAndReturn(
			func() ([]protocol.Endpoint, error) {
				return availableEndpoints, nil
			},
		).AnyTimes()


	mockProtocol := gatewaymocks.NewMockProtocol(ctrl)
	mockProtocol.EXPECT().
		// PARAM: Service ID
		// PARAM: HTTP Request
		BuildRequestContext(gomock.Any(), gomock.Any()).
		Return(mockReqCtx, nil).
		AnyTimes()

	return mockProtocol
}

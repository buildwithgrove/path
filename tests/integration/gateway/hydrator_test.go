package integrationtest

import (
	"testing"

//	"github.com/golang/mock/gomock"
//	"github.com/pokt-network/poktroll/pkg/polylog/polyzero"
        "github.com/stretchr/testify/require"

	"github.com/buildwithgrove/path/testutil/gateway"
)

// This is an integratio test for the EndpointHydrator + Solana Service QoS
// It checks whether the combination of the two can successfully filter out invalid endpoints.
// A mock protocol instance is used to return the endpoint responses required to construct the test scenario.
func TestEndpointHydrator_Solana(t *testing.T) {

	// setup Solana QoS instance
	// setup EndpointHydrator instance
	// scenarios:
	//	A. 2 valid endpoints from a group of 4.
	//	B. all invalid endpoints
	//      C. invalid reason: getHealth
	//      D. invalid reason: getEpochInfo

	protocol := gateway.Protocol(t)

	protocolReqCtx, err := protocol.BuildRequestContext("", nil)
	require.NoError(t, err)

	eps, err := protocolReqCtx.AvailableEndpoints()
	require.NoError(t, err)
	require.Equal(t, 4, len(eps))
}


/*
func setupEndpointHydrator(t *testing.T) ( {
	t.Helper()

	logger := polyzero.NewLogger()

}
*/



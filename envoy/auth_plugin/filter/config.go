//go:build auth_plugin

package filter

import (
	xds "github.com/cncf/xds/go/xds/type/v3"
	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
	"google.golang.org/protobuf/types/known/anypb"
)

type EnvoyConfig struct {
	echoBody string
}

type Parser struct{}

// Parse the filter configuration. We can call the ConfigCallbackHandler to control the filter's
// behavior
func (p *Parser) Parse(any *anypb.Any, callbacks api.ConfigCallbackHandler) (interface{}, error) {
	configStruct := &xds.TypedStruct{}
	if err := any.UnmarshalTo(configStruct); err != nil {
		return nil, err
	}
	return &EnvoyConfig{}, nil
}

// Merge configuration from the inherited parent configuration
func (p *Parser) Merge(parent interface{}, child interface{}) interface{} {
	parentConfig := parent.(*EnvoyConfig)
	childConfig := child.(*EnvoyConfig)

	// copy one, do not update parentConfig directly.
	newConfig := *parentConfig
	if childConfig.echoBody != "" {
		newConfig.echoBody = childConfig.echoBody
	}
	return &newConfig
}

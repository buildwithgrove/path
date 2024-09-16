//go:build auth_plugin

package filter

import (
	xds "github.com/cncf/xds/go/xds/type/v3"
	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
	"google.golang.org/protobuf/types/known/anypb"
)

type envoyConfig struct {
	echoBody string
}

type parser struct {
}

// Parse the filter configuration. We can call the ConfigCallbackHandler to control the filter's
// behavior
func (p *parser) Parse(any *anypb.Any, callbacks api.ConfigCallbackHandler) (interface{}, error) {
	configStruct := &xds.TypedStruct{}
	if err := any.UnmarshalTo(configStruct); err != nil {
		return nil, err
	}

	// v := configStruct.Value
	conf := &envoyConfig{}
	// prefix, ok := v.AsMap()["prefix_localreply_body"]
	// if !ok {
	// 	return nil, errors.New("missing prefix_localreply_body")
	// }
	// if str, ok := prefix.(string); ok {
	// 	conf.echoBody = str
	// } else {
	// 	return nil, fmt.Errorf("prefix_localreply_body: expect string while got %T", prefix)
	// }
	return conf, nil
}

// Merge configuration from the inherited parent configuration
func (p *parser) Merge(parent interface{}, child interface{}) interface{} {
	parentConfig := parent.(*envoyConfig)
	childConfig := child.(*envoyConfig)

	// copy one, do not update parentConfig directly.
	newConfig := *parentConfig
	if childConfig.echoBody != "" {
		newConfig.echoBody = childConfig.echoBody
	}
	return &newConfig
}

func filterFactory(c interface{}, callbacks api.FilterCallbackHandler) api.StreamFilter {
	conf, ok := c.(*envoyConfig)
	if !ok {
		panic("unexpected config type")
	}
	return &filter{
		callbacks: callbacks,
		config:    conf,
	}
}

func main() {}

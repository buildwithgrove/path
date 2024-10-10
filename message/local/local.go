// package local provides the publishing functionality
// required to skip sharing data between PATH instances,
// by informing the local PATH components instead.
// This allows the running of a single PATH instance
// without the need for setting up a messaging platform,
// e.g. NATS, REDIS.
package local

import (
	"fmt"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/message"
)

// Publisher provides the functionality required by the
// gateway package to publish an ObservationSet.
var _ gateway.QoSPublisher = &Publisher{}

type Publisher struct{}

func (p Publisher) Publish(observationSet message.ObservationSet) error {
	err := observationSet.Broadcast()
	if err != nil {
		return fmt.Errorf("local publisher: failed to broadcast observation set: %w", err)
	}

	return nil
}

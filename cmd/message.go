package main

import (
	"github.com/buildwithgrove/path/config"
	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/message/local"
)

// getQoSPublisher returns an instance of publisher
// which meets the requirements of the gateway package
// for QoSPublisher.
// The returned publisher instance is determined by the
// messaging system's config.
// e.g. return a "local" publisher if no messaging systems
// have been configured.
func getQoSPublisher(config config.MessagingConfig) (gateway.QoSPublisher, error) {
	// TODO_UPNEXT(@adshmh): return a publisher that matches the
	// configuration settings of the messaging system, once a
	// pub/sub messaging platform has been selected.
	//
	// Return a Local Publisher instance if no messaging system(s) are
	// configured. A Local Publisher only informs the local components
	// when asked to publish an Observation Set.
	return local.Publisher{}, nil
}

package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/gateway"
	"github.com/buildwithgrove/path/observation"
)

// DataReporterHTTP exports observations to an external components over HTTP (e.g. Flentd HTTP Plugin, a Messaging system, or a database)
var _ gateway.RequestResponseReporter = &DataReporterHTTP{}

// DataReporterHTTP sends the observation for each handled request to an HTTP endpoint.
// It assumes the HTTP server is part of the data pipeline, i.e. it processes and stores/forwards the observations as appropriate.
// For example: a Fluentd HTTP input plugin, with output plugin pointing to BigQuery.
// Implements the gateway.RequestResponseReporter
type DataReporterHTTP struct {
	Logger polylog.Logger

	// The URL of the Data Pipeline's HTTP server.
	// Only JSON-accepting data pipelines are supported as of PR #215.
	// e.g. Fluentd HTTP input plugin on localhost:8686.
	DataProcessorURL string
}

// Publish the supplied observations:
// - Build the expected data record.
// - Send to the configured URL.
func (drh *DataReporterHTTP) Publish(observations *observation.RequestResponseObservations) {
	logger := drh.hydrateLogger(observations)

	// TODO_MVP(@adshmh): Replace this with the new DataRecord struct once the data pipeline is updated.
	// convert to legacy-formatted data record
	legacyDataRecord := buildLegacyDataRecord(logger, observations)

	// Marshal the data record.
	serializedRecord, err := json.Marshal(legacyDataRecord)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to serialize the data record. Skip reporting.")
		return
	}

	// Send the marshaled data record to the data processor.
	if err := drh.sendRecordOverHTTP(serializedRecord); err != nil {
		logger.Warn().Err(err).Msg("Failed to send the data record over HTTP. Skip reporting.")
	}
}

func (drh *DataReporterHTTP) sendRecordOverHTTP(serializedDataRecord []byte) error {
	// Send the marshaled bytes to the data processor, e.g. Fluentd.
	//
	resp, err := http.Post(drh.DataProcessorURL, "application/json", bytes.NewReader(serializedDataRecord))
	if err != nil {
		return err
	}

	// Verify the data processor responded with OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error sending the data record: got HTTP status %d, expected %d", resp.StatusCode, http.StatusOK)
	}

	return nil
}

// hydrateLogger enhances the logger with observation data:
// - Starts with component and service info
// - Adds gateway data if available
// - Adds auth data if available
func (drh *DataReporterHTTP) hydrateLogger(observations *observation.RequestResponseObservations) polylog.Logger {
	// Base logger with component and service ID
	logger := drh.Logger.With(
		"component", "DataReporterHTTP",
		"service_id", observations.ServiceId,
	)

	gatewayObservations := observations.GetGateway()
	// Skip if no gateway observations
	if gatewayObservations == nil {
		return logger
	}

	// Add request type (user/hydrator)
	logger = logger.With("request_type", gatewayObservations.GetRequestType().String())

	requestAuth := gatewayObservations.GetRequestAuth()
	// Skip if no auth data
	if requestAuth == nil {
		return logger
	}

	// Add request ID for tracing
	logger = logger.With("trace_id", requestAuth.GetTraceId())
	return logger
}
